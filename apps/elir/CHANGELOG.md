# Changelog

Elir's versions, newest first; up to 0.2.0 the app was called Kit. `scripts/release.sh` takes the section of the version it releases as the release notes.

## 0.5.0

- Dictation writes for the app in front. In a terminal or an AI app, such as Claude Code, Codex or ChatGPT, file names,
  commands and other code come out spelled exactly, in backticks; in a chat, such as WeChat, Telegram or Discord, the
  text reads as a message you typed, without a full stop at the end. Turn it off with "Format for the app in front" in
  Dictation & Translation.

## 0.4.3

- Dictation keeps what you said when "actually", 其实 or 不是…而是… only stress or contrast rather than correct
  yourself, leaves short or already clean dictation as it is, and makes a list only of items you go through one by one.
  A dictation that tells the model to ignore its rules is written down, not obeyed.
- Polishing selected text keeps it in its own language: Chinese, Japanese or Korean text is no longer translated, even
  in part.

## 0.4.2

- Dictation cleans up Chinese better: it keeps only the final version when you correct yourself, fixes words speech
  recognition misheard from what you are talking about, splits long dictation into paragraphs, and keeps your
  questions, your own words and every detail instead of rewording them in a formal style. Translation keeps names and
  times it used to drop.

## 0.4.1

- A tap of Fn starts dictation again; since 0.4.0 the recording stopped as it began, as "Too short, ignored".
- Shortcuts are written as macOS writes them: `⌥T` and `⇧⌘2` rather than "Alt + T" and "Command + Shift + 2".

## 0.4.0

- Every shortcut can be changed: click one in Dictation & Translation, Translate & Polish or Screenshot, press the new
  keys and save. A shortcut can tell left from right modifiers, and Elir says when one conflicts with another. The
  screenshot's `Space` and `Esc` stay those of macOS's own capture.
- Quitting Elir while it records puts the sound it muted or lowered back first, and a muted output device that is no
  longer there no longer keeps later recordings from quieting the sound.
- A settings or API key file that Elir cannot make sense of is kept aside, as `settings.json.unreadable-…` or
  `secrets.json.unreadable-…`, and Elir starts afresh instead of failing to save.
- Elir also runs on Windows, built from source for now: see the README.

## 0.3.0

- Translate and polish selected text: select text in any app and press `⌥T` to translate it, or `⌥R` to polish it.
  The result replaces the selection (`⌘Z` undoes it); `Esc` cancels. Translation follows the languages of Dictation &
  Translation, and both use the model and API key of Models. See Settings → Translate & Polish.

## 0.2.4

- API keys are kept in a file of Elir's data that only you can read, instead of the Keychain, which asked for your
  login password after every update. Moving them may ask one last time.

## 0.2.3

- The install steps clear the mark macOS puts on the downloaded disk image first, so that it opens. The app is as in
  0.2.2.

## 0.2.2

- The menu bar icon is now the head of the character of Elir's icon, tinted to match a light or dark menu bar.

## 0.2.1

- Kit is now **Elir**, with a new icon. Settings, history, screenshots, API keys and permissions carry over. An
  installed copy that updates itself keeps its file name, Kit.app, which can be renamed in Finder.
- Settings → General → Show in Dock gives Elir a Dock icon, and a place in ⌘Tab, besides its menu bar icon.

## 0.2.0

- Kit updates itself: Settings → General checks for a new version, shows what's new and installs it with one click,
  then relaunches. It checks by itself once a day, which the same page can turn off; the menu bar menu shows an update
  that is waiting.
- Coming from 0.1.0, download this version once by hand: 0.1.0 cannot update itself.

## 0.1.0

The first release of Kit, a menu bar toolbox for macOS.

### Voice input

- `Fn` dictates: speech is transcribed, cleaned up and pasted where the cursor is.
- `Fn` + `Left Shift` translates instead; `Tab` or a click on the voice bar switches the target language, `Esc`
  cancels.
- Soniox transcribes while you speak, with OpenRouter as the fallback; OpenRouter cleans up and translates.
- Music and videos are muted or lowered while recording.
- A dictionary of names and terms, and a searchable history of the text, which can be turned off.

### Screenshots

- `⌘⇧2` captures an area or, with `Space`, a window, as macOS's `⌘⇧4` does.
- The screenshot is copied to the clipboard (paste with `Control+V` in Claude Code and other terminals), and its
  thumbnail can be dragged into any app.
- Screenshots are kept in Kit, never on the Desktop, for 1, 7 or 30 days or forever, and listed in Settings.

### The app

- English and Simplified Chinese, light and dark, and an accent color.
- Open at login.
- Brings over the settings, history and API keys of Voice, the app Kit grew from.
