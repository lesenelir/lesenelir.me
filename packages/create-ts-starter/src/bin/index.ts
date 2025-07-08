#!/usr/bin/env node
import { existsSync, readdirSync } from 'node:fs'
import fsPromises from 'node:fs/promises'
import path from 'node:path'

import fse from 'fs-extra'
import { bold, green, red } from 'kolorist'
import prompts from 'prompts'

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

function canSafelyOverwrite(dir: string): boolean {
  const directoryExists = existsSync(dir)

  if (!directoryExists) {
    return true
  }

  // the file exists, but is it empty?
  // when it's empty, we can safely overwrite it
  return readdirSync(dir).length === 0
}

async function main() {
  let targetDir = ''
  const defaultProjectName = 'ts-starter'

  let result: {
    packageName: string
    shouldOverwrite: string
  }

  try {
    result = await prompts(
      [
        {
          type: 'text',
          name: 'projectName',
          message: 'What is your project named?',
          initial: defaultProjectName,
          onState: (state) => (targetDir = String(state.value).trim() || defaultProjectName)
        },
        {
          type: () => (canSafelyOverwrite(targetDir) ? null : 'confirm'),
          name: 'shouldOverwrite',
          message: `${targetDir} is not empty. Remove existing files and continue?`
        },
        {
          name: 'overwriteChecker',
          type: (values) => {
            if (values === false) {
              throw new Error(red('✖') + ' Operation cancelled')
            }
            return null
          }
        },
        {
          type: () => (isValidPackageName(targetDir) ? null : 'text'),
          name: 'packageName',
          message: 'Package name',
          initial: () => toValidPackageName(targetDir),
          validate: (dir: string) => isValidPackageName(dir) || 'Invalid package.json name'
        }
      ],
      {
        onCancel: () => {
          throw new Error(red('✖') + ' Operation cancelled')
        }
      }
    )
  } catch (e) {
    if (e instanceof Error) {
      console.log(e.message)
    }
    process.exit(1)
  }

  const { packageName, shouldOverwrite } = result
  const root = path.resolve(targetDir)

  if (shouldOverwrite) {
    fse.emptyDirSync(root)
  } else if (!existsSync(root)) {
    await fsPromises.mkdir(root, { recursive: true })
  }

  const pkg = {
    name: packageName ?? toValidPackageName(targetDir),
    version: '0.0.0'
  }

  console.log('Setting up project ...')

  // -----------------------------------------------------------------------------------
  const templateDir = path.join(__dirname, '../../template')

  // read existing package.json from the root directory
  const packageJsonPath = path.join(root, 'package.json')

  // read new package.json from the template directory
  const templatePackageJsonPath = path.join(templateDir, 'package.json')
  const templatePackageJson = JSON.parse(
    await fsPromises.readFile(templatePackageJsonPath, 'utf-8')
  )

  fse.copySync(templateDir, root)

  if (existsSync(path.join(root, 'gitignore'))) {
    await fsPromises.rename(path.join(root, 'gitignore'), path.join(root, '.gitignore'))
  }

  await fsPromises.writeFile(
    packageJsonPath,
    JSON.stringify(
      {
        ...templatePackageJson,
        ...pkg
      },
      null,
      2
    )
  )

  // -----------------------------------------------------------------------------------
  const manager = process.env.npm_config_user_agent ?? ''
  const packageManager = /pnpm/.test(manager) ? 'pnpm' : /yarn/.test(manager) ? 'yarn' : 'npm'

  const commandsMap = {
    install: {
      pnpm: 'pnpm install',
      yarn: 'yarn',
      npm: 'npm install'
    },
    dev: {
      pnpm: 'pnpm start',
      yarn: 'yarn start',
      npm: 'npm run start'
    }
  }

  console.log(`\nDone. Now run:\n`)
  console.log(`${bold(green(`cd ${targetDir}`))}`)
  console.log(`${bold(green(commandsMap.install[packageManager]))}`)
  console.log(`${bold(green(commandsMap.dev[packageManager]))}`)
  console.log()
}

main().then(() => {})
