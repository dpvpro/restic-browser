<img src="./frontend/src/assets/images/eye.png" alt="drawing" height="48px"/> <img src="./frontend/src/assets/images/logo.png" alt="drawing" height="48px"/> 

---

A simple, cross-platform [restic backup](https://github.com/restic/restic) GUI for browsing and restoring restic repositories. 

Built with [Wails3](https://v3.wails.io), based on leaanthony's [Restoric](https://github.com/leaanthony/restoric) PoC. 


## Download

Prebuilt binaries can be downloaded from the [GitHub releases](https://github.com/emuell/restic-browser/releases) page.


## Features

* *Displays* contents (snapshots, files) from local and remote restic repositories.
* *Dumps* selected files or folders (as a zip archive) to a desired location.
* *Restores* selected files or folders to a desired location.
* *Opens* selected files by moving them to TEMP, then opens them with your operating system's default programs.

This is not a fullblown restic backup GUI - it only allows you to *browse* existing repositories!  

![Screenshot](./screenshot.png "Restic Browser")


## Keyboard Navigation

The UI is navigatable via keyboard shortcuts. To change the focus area, hit `Tab` + `Shift-Tab` keys.

### Global Shortcuts: 

- `Control/CMD + O`: Open new repository

### Snapshot-List
- `Arrow keys`, `Page Up/Down`, `Home/End`: Change selected snapshot

### File-List
- `Arrow keys`, `Page Up/Down`, `Home/End`: Change selected file
- `o` or `Enter` or `Space`: Open selected file or folder
- `d`: Dump selected file or folder as zip file
- `r`: Restore selected file or folder


## System Requirements

#### All platforms
- Install [restic](https://github.com/restic/restic/releases/) and *make sure it is included in your $PATH* or 
pass the path to the restic executable *as argument*: `Restic-Browser[.exe] --restic="PATH/TO/restic[.exe]"`.

#### Windows:
- Windows 10 or later with [WebView2 Runtime](https://developer.microsoft.com/de-de/microsoft-edge/webview2/#download-section)
#### MacOS:
- macOS 10.14 or later
#### Linux:
- Linux with GLIBC_2.31 or later (e.g. Ubuntu 20.04 or later)
- GTK4 and WebKitGTK 6.0 (install via `apt install libgtk-4-dev libwebkitgtk-6.0-dev` on Ubuntu)


## Development

### Dependencies

* Follow the [Wails3 Installation Docs](https://v3.wails.io/getting-started/installation/) to install the `wails3` CLI for your platform.
* Install [restic](https://github.com/restic/restic/releases/) and *make sure it is included in your $PATH*. 

### Front-end and App Development

To work in live development mode with automatic hot-reloading, run `wails3 dev` in the root directory. 

### Go Backend Debugging

To debug the Wails Go application code, you can use the included startup tasks of vscode. If you press "F5" in vscode, the application will be built in debug mode and then started.   

### Building Production Packages

To build a redistributable package in production mode, use `wails3 build` (the binary is written to `bin/`).
To create platform packages, use `wails3 package`.


## License

MIT license. See [LICENSE](./LICENSE) for the full text.


## Contribute

Patches are welcome! Please fork the latest git repository and create a feature branch. 
