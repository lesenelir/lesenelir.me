#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import { existsSync, readdirSync } from 'node:fs'
import fs from 'node:fs/promises'
import path from 'node:path'
import { parseArgs, styleText } from 'node:util'
import * as p from '@clack/prompts'

const TEMPLATE_DIR = path.resolve(import.meta.dirname, '../../template')
const DEFAULT_DIR = 'ts-starter'

const HELP = `Usage: create-ts-starter [dir] [options]

Options:
  --name <name>  package name (defaults to the directory name)
  --overwrite    empty a non-empty directory without asking
  --no-git       don't run git init
  --no-install   don't install dependencies
  -y, --yes      take the defaults instead of asking
  -h, --help     show this help
`

type PackageManager = 'pnpm' | 'yarn' | 'bun' | 'npm'

// check if the package name is valid
// unValid package name: MyProject | @my-scope/ | -my-project
function isValidPackageName(projectName: string): boolean {
  return /^(?:@[a-z0-9-*~][a-z0-9-*._~]*\/)?[a-z0-9-~][a-z0-9-._~]*$/.test(projectName)
}

function toValidPackageName(projectName: string) {
  return projectName
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/^[._]/, '')
    .replace(/[^a-z0-9-~]+/g, '-')
}

// a directory holding nothing but .git can be filled without asking
function isEmpty(dir: string): boolean {
  return !existsSync(dir) || readdirSync(dir).every(file => file === '.git')
}

async function emptyDir(dir: string) {
  for (const file of await fs.readdir(dir)) {
    if (file !== '.git') {
      await fs.rm(path.join(dir, file), { recursive: true, force: true })
    }
  }
}

function isInsideGitRepo(dir: string): boolean {
  let existing = dir
  while (!existsSync(existing)) {
    existing = path.dirname(existing)
  }
  return spawnSync('git', ['rev-parse', '--is-inside-work-tree'], { cwd: existing, stdio: 'ignore' }).status === 0
}

function detectPackageManager(): PackageManager {
  const name = (process.env.npm_config_user_agent ?? '').split('/')[0]
  return name === 'pnpm' || name === 'yarn' || name === 'bun' ? name : 'npm'
}

function runCommand(pm: PackageManager, script: string): string {
  return pm === 'npm' || pm === 'bun' ? `${pm} run ${script}` : `${pm} ${script}`
}

function cancel(message = 'Operation cancelled'): never {
  p.cancel(message)
  process.exit(1)
}

// exits when the prompt was cancelled with Ctrl+C or Esc
function answer<T>(value: T): Exclude<T, symbol> {
  if (p.isCancel(value)) {
    cancel()
  }
  return value as Exclude<T, symbol>
}

async function main() {
  const { values, positionals } = parseArgs({
    allowPositionals: true,
    allowNegative: true,
    options: {
      name: { type: 'string' },
      overwrite: { type: 'boolean' },
      git: { type: 'boolean' },
      install: { type: 'boolean' },
      yes: { type: 'boolean', short: 'y' },
      help: { type: 'boolean', short: 'h' },
    },
  })

  if (values.help) {
    console.log(HELP)
    return
  }

  const yes = values.yes ?? false
  const pm = detectPackageManager()

  p.intro(styleText('inverse', ' create-ts-starter '))

  // -----------------------------------------------------------------------------------
  const targetDir =
    (
      positionals[0] ??
      (yes
        ? DEFAULT_DIR
        : answer(
            await p.text({
              message: 'Where should the project be created?',
              placeholder: DEFAULT_DIR,
              defaultValue: DEFAULT_DIR,
            })
          ))
    ).trim() || DEFAULT_DIR
  const root = path.resolve(targetDir)

  if (!isEmpty(root)) {
    const overwrite =
      values.overwrite ??
      (!yes &&
        answer(
          await p.confirm({
            message: `${targetDir} is not empty. Remove its files and continue?`,
            initialValue: false,
          })
        ))
    if (!overwrite) {
      cancel(yes ? `${targetDir} is not empty; pass --overwrite to empty it` : undefined)
    }
  }

  const dirName = path.basename(root)
  const packageName =
    values.name ??
    (isValidPackageName(dirName)
      ? dirName
      : yes
        ? toValidPackageName(dirName)
        : answer(
            await p.text({
              message: 'Package name',
              initialValue: toValidPackageName(dirName),
              validate: value => (isValidPackageName(value ?? '') ? undefined : 'Invalid package.json name'),
            })
          ))
  if (!isValidPackageName(packageName)) {
    cancel(`Invalid package.json name: ${packageName}`)
  }

  // a project created inside another repository belongs to it, so don't start a new one
  const git =
    values.git ??
    (!isInsideGitRepo(root) && (yes || answer(await p.confirm({ message: 'Initialize a git repository?' }))))

  const install = values.install ?? (yes || answer(await p.confirm({ message: `Install dependencies with ${pm}?` })))

  // -----------------------------------------------------------------------------------
  const spinner = p.spinner()
  spinner.start('Copying the template')

  if (existsSync(root)) {
    await emptyDir(root)
  } else {
    await fs.mkdir(root, { recursive: true })
  }

  await fs.cp(TEMPLATE_DIR, root, { recursive: true })

  // npm never publishes a .gitignore, so the template ships it as _gitignore
  if (existsSync(path.join(root, '_gitignore'))) {
    await fs.rename(path.join(root, '_gitignore'), path.join(root, '.gitignore'))
  }

  const packageJsonPath = path.join(root, 'package.json')
  const templatePackageJson = JSON.parse(await fs.readFile(packageJsonPath, 'utf-8'))
  // the repository url is named after the package, without its scope
  const repoName = packageName.split('/').pop() ?? packageName
  const packageJson = {
    ...templatePackageJson,
    name: packageName,
    version: '0.0.0',
    repository: {
      ...templatePackageJson.repository,
      url: String(templatePackageJson.repository?.url ?? '').replace('[project-name]', repoName),
    },
  }
  // corepack would refuse to run another package manager in a project pinned to pnpm
  if (pm !== 'pnpm') {
    delete packageJson.packageManager
  }
  await fs.writeFile(packageJsonPath, `${JSON.stringify(packageJson, null, 2)}\n`)

  await fs.writeFile(path.join(root, 'README.md'), `# ${packageName}\n`)

  const licensePath = path.join(root, 'LICENSE')
  if (existsSync(licensePath)) {
    const license = await fs.readFile(licensePath, 'utf-8')
    await fs.writeFile(
      licensePath,
      license.replace(/Copyright \(c\) \d{4}/, `Copyright (c) ${new Date().getFullYear()}`)
    )
  }

  spinner.stop(`Created ${packageName} in ${root}`)

  if (pm !== 'pnpm') {
    p.log.warn(`The GitHub workflows in .github use pnpm; change them to ${pm} before pushing.`)
  }

  // -----------------------------------------------------------------------------------
  if (git && spawnSync('git', ['init', '-q'], { cwd: root, stdio: 'ignore' }).status !== 0) {
    p.log.warn('git init failed; run it yourself.')
  }

  let installed = false
  if (install) {
    p.log.step(`Installing dependencies with ${pm}`)
    installed =
      spawnSync(pm, ['install'], { cwd: root, stdio: 'inherit', shell: process.platform === 'win32' }).status === 0
    if (!installed) {
      p.log.error('Installing dependencies failed; run the install yourself.')
    }
  }

  const relativeDir = path.relative(process.cwd(), root)
  const steps = [
    relativeDir && `cd ${relativeDir.includes(' ') ? `"${relativeDir}"` : relativeDir}`,
    !installed && (pm === 'yarn' ? 'yarn' : `${pm} install`),
    runCommand(pm, 'dev'),
  ].filter(Boolean)

  p.note(steps.join('\n'), 'Next steps')
  p.outro(styleText('green', 'Done'))
}

main().catch((error: unknown) => {
  cancel(error instanceof Error ? error.message : String(error))
})
