package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"restic-browser/backend/open"
	"restic-browser/backend/restic"

	"github.com/leaanthony/sail/fs"
	"github.com/pkg/errors"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/text/encoding/unicode"
)

type ResticBrowserApp struct {
	tempPath    string
	restic      *restic.Restic
	resticError error
	repo        *restic.Repository
	snapshots   map[string]*restic.Snapshot
}

func NewResticBrowser(resticPath *string) *ResticBrowserApp {
	program, err := restic.NewRestic(resticPath)
	if err != nil {
		// continue without a valid restic instance
		fmt.Print(err.Error() + "\n")
	}
	return &ResticBrowserApp{
		restic:      program,
		resticError: err,
		snapshots:   make(map[string]*restic.Snapshot),
	}
}

func (r *ResticBrowserApp) showWarning(title, message string) {
	r.showMessage(title, message, application.WarningDialogType, nil)
}

func (r *ResticBrowserApp) showError(title, message string) {
	r.showMessage(title, message, application.ErrorDialogType, nil)
}

// showConfirmation asks the user a question and blocks until one of the buttons
// has been clicked.
func (r *ResticBrowserApp) showConfirmation(title, message string) bool {
	result := make(chan bool, 1)
	dialog := application.Get().Dialog.Question()
	dialog.SetTitle(title)
	dialog.SetMessage(message)
	dialog.AddButton("Yes").OnClick(func() {
		result <- true
	})
	noButton := dialog.AddButton("No")
	noButton.OnClick(func() {
		result <- false
	})
	dialog.SetCancelButton(noButton)
	dialog.Show()
	return <-result
}

// showMessage displays a message dialog. If onClose is not nil, it is called
// when the user dismisses the dialog.
func (r *ResticBrowserApp) showMessage(title, message string, messageType application.DialogType, onClose func()) {
	app := application.Get()
	var dialog *application.MessageDialog
	switch messageType {
	case application.ErrorDialogType:
		dialog = app.Dialog.Error()
	case application.QuestionDialogType:
		dialog = app.Dialog.Question()
	default:
		dialog = app.Dialog.Warning()
	}
	dialog.SetTitle(title)
	dialog.SetMessage(message)
	button := dialog.AddButton("Okay")
	if onClose != nil {
		button.OnClick(onClose)
	}
	dialog.Show()
}

func (r *ResticBrowserApp) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	// create app temp dir
	tempPath, err := os.MkdirTemp(os.TempDir(), "restic-browser")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create app temp dir: %s\n", err.Error())
	}
	r.tempPath = tempPath
	return nil
}

func (r *ResticBrowserApp) DomReady() {
	// warn about missing restic program: this is the first time we can show a dialog
	if r.restic == nil {
		message := fmt.Sprintf(
			"Failed to find restic program: %s\n\n", r.resticError.Error()) +
			"Please select your installed restic binary manually in the following dialog."
		r.showMessage("Restic Binary Missing", message, application.WarningDialogType, r.selectResticBinary)
	}
}

// selectResticBinary asks the user to locate the restic binary manually.
func (r *ResticBrowserApp) selectResticBinary() {
	path, err := application.Get().Dialog.OpenFile().
		SetTitle("Please select your restic program").
		ShowHiddenFiles(true).
		TreatsFilePackagesAsDirectories(true).
		CanChooseFiles(true).
		CanChooseDirectories(false).
		PromptForSingleSelection()
	if err == nil && path != "" {
		r.restic, err = restic.NewResticFromPath(path)
		if err != nil {
			r.showError("Restic Binary Error",
				fmt.Sprintf("Failed to set restic binary: %s", err.Error()))
		}
	}
}

func (r *ResticBrowserApp) ServiceShutdown() error {
	// remove app temp dir
	err := os.RemoveAll(r.tempPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove app temp dir: %s\n", err.Error())
	}
	return nil
}

func (r *ResticBrowserApp) OpenFileOrUrl(path string) error {
	return open.OpenFileOrURL(path)
}

// read a text file, skipping BOM headers if present
func readTextFile(filename string) ([]byte, error) {
	var (
		bomUTF8              = []byte{0xef, 0xbb, 0xbf}
		bomUTF16BigEndian    = []byte{0xfe, 0xff}
		bomUTF16LittleEndian = []byte{0xff, 0xfe}
	)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(data, bomUTF8) {
		return data[len(bomUTF8):], nil
	}
	if !bytes.HasPrefix(data, bomUTF16BigEndian) && !bytes.HasPrefix(data, bomUTF16LittleEndian) {
		return data, nil
	}
	// UseBom means automatic endianness selection
	e := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	return e.NewDecoder().Bytes(data)
}

// defaultRepo determines the default restic repository location from the environment.
func defaultRepo() (string, error) {
	repositoryFile := os.Getenv("RESTIC_REPOSITORY_FILE")
	if repositoryFile != "" {
		s, err := readTextFile(repositoryFile)
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%s does not exist", repositoryFile)
		}
		return strings.TrimSpace(strings.Replace(string(s), "\n", "", -1)), err
	}
	repository := os.Getenv("RESTIC_REPOSITORY")
	return repository, nil
}

// defaultRepoPassword determines the default restic repository password from the environment.
func defaultRepoPassword() (string, error) {
	passwordFile := os.Getenv("RESTIC_PASSWORD_FILE")
	if passwordFile != "" {
		s, err := readTextFile(passwordFile)
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%s does not exist", passwordFile)
		}
		return strings.TrimSpace(strings.Replace(string(s), "\n", "", -1)), err
	}
	password := os.Getenv("RESTIC_PASSWORD")
	return password, nil
}

func (r *ResticBrowserApp) DefaultRepoLocation() restic.Location {
	location := restic.Location{}
	location.Path, _ = defaultRepo()
	location.Password, _ = defaultRepoPassword()
	if len(location.Path) == 0 {
		// skip reading other location info when no repo is present
		return location
	}
	type LocationInfo struct {
		prefix      string
		credentials []string
	}
	locationInfos := []LocationInfo{
		{prefix: "bs"},
		{prefix: "sftp"},
		{prefix: "rest"},
		{prefix: "rclone"},
		{prefix: "s3", credentials: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}},
		{prefix: "b2", credentials: []string{"B2_ACCOUNT_ID", "B2_ACCOUNT_KEY"}},
		{prefix: "azure", credentials: []string{"AZURE_ACCOUNT_NAME", "AZURE_ACCOUNT_KEY"}},
	}
	location.Prefix = ""
	for _, locationInfo := range locationInfos {
		if strings.HasPrefix(location.Path, locationInfo.prefix+":") {
			location.Prefix = locationInfo.prefix
			location.Path = strings.Replace(location.Path, locationInfo.prefix+":", "", 1)
			for _, credential := range locationInfo.credentials {
				location.Credentials = append(location.Credentials,
					restic.EnvValue{Name: credential, Value: os.Getenv(credential)})
			}
			break
		}
	}
	return location
}

func (r *ResticBrowserApp) ReadPasswordFromFile() (string, error) {
	filename, err := application.Get().Dialog.OpenFile().
		SetTitle("Please select a password file").
		ShowHiddenFiles(true).
		ResolvesAliases(true).
		TreatsFilePackagesAsDirectories(true).
		CanCreateDirectories(false).
		PromptForSingleSelection()
	if err == nil && filename == "" {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to show dialog: %s", err.Error())
	}
	bytes, err := readTextFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %s", err.Error())
	}
	return strings.TrimSpace(strings.Replace(string(bytes), "\n", "", -1)), nil
}

func (r *ResticBrowserApp) SelectLocalRepo() (string, error) {
	dir, err := application.Get().Dialog.OpenFile().
		SetTitle("Please select a restic repository directory").
		ShowHiddenFiles(true).
		ResolvesAliases(true).
		TreatsFilePackagesAsDirectories(true).
		CanCreateDirectories(false).
		CanChooseFiles(false).
		CanChooseDirectories(true).
		PromptForSingleSelection()
	if err == nil && dir == "" {
		return "", nil
	}
	if err != nil || !restic.IsDirectoryARepository(dir) {
		return "", fmt.Errorf("directory doesn't look like a restic backup location")
	}
	return dir, nil
}

func (r *ResticBrowserApp) OpenRepo(location restic.Location) ([]*restic.Snapshot, error) {
	if r.restic == nil {
		return nil, fmt.Errorf("failed to find restic program")
	}
	repo := restic.NewRepository(location, r.restic)
	snapshots, err := repo.GetSnapshots()
	if err != nil {
		return nil, err
	}
	r.repo = repo

	// Save snapshot details
	for _, snapshot := range snapshots {
		r.snapshots[snapshot.ID] = snapshot
	}

	return snapshots, nil
}

func (r *ResticBrowserApp) GetFilesForPath(snapshotID, path string) ([]*restic.File, error) {
	snapshot := r.snapshots[snapshotID]
	if snapshot == nil {
		return nil, fmt.Errorf("%s is not a valid snapshot ID", snapshotID)
	}
	return r.repo.GetFiles(snapshot, path)
}

func (r *ResticBrowserApp) RestoreFile(snapshotID string, file *restic.File) (string, error) {
	snapshot := r.snapshots[snapshotID]
	if snapshot == nil {
		return "", fmt.Errorf("%s is not a valid snapshot ID", snapshotID)
	}
	options := application.OpenFileDialogOptions{
		Title:                           "Please select a target directory",
		ShowHiddenFiles:                 true,
		CanCreateDirectories:            true,
		ResolvesAliases:                 true,
		TreatsFilePackagesAsDirectories: true,
		CanChooseFiles:                  false,
		CanChooseDirectories:            true,
	}
	targetPath, err := application.Get().Dialog.OpenFileWithOptions(&options).PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if targetPath == "" {
		return "", nil
	}

	targetFileOrFolderName := filepath.Join(targetPath, file.Path)
	if file.Type == "dir" {
		if fs.DirExists(targetFileOrFolderName) {
			if !r.showConfirmation("Overwrite existing directory?",
				fmt.Sprintf("The target directory '%s' already exists.\n\n"+
					"Are you sure that you want to overwrite the existing directory?",
					targetFileOrFolderName)) {
				return "", fmt.Errorf("target directory '%s' already exists", targetFileOrFolderName)
			}
		}
	} else {
		if fs.FileExists(targetFileOrFolderName) {
			if !r.showConfirmation("Overwrite existing file?",
				fmt.Sprintf("The target file '%s' already exists.\n\n"+
					"Are you sure that you want to overwrite the existing file?",
					targetFileOrFolderName)) {
				return "", fmt.Errorf("target directory '%s' already exists", targetFileOrFolderName)
			}
		}
	}

	targetFilePath, err := r.repo.RestoreFile(snapshot, file, targetPath)
	if err != nil {
		return "", err
	}
	return targetFilePath, nil
}

func (r *ResticBrowserApp) DumpFile(snapshotID string, file *restic.File) (string, error) {
	snapshot := r.snapshots[snapshotID]
	if snapshot == nil {
		return "", fmt.Errorf("%s is not a valid snapshot ID", snapshotID)
	}
	options := application.OpenFileDialogOptions{
		Title:                           "Please select a target directory",
		ShowHiddenFiles:                 true,
		CanCreateDirectories:            true,
		ResolvesAliases:                 true,
		TreatsFilePackagesAsDirectories: true,
		CanChooseFiles:                  false,
		CanChooseDirectories:            true,
	}
	targetPath, err := application.Get().Dialog.OpenFileWithOptions(&options).PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if targetPath == "" {
		return "", nil
	}

	targetFileName := filepath.Join(targetPath, file.Name)
	if file.Type == "dir" {
		targetFileName += ".zip"
	}
	if fs.FileExists(targetFileName) {
		if !r.showConfirmation("Overwrite existing file?",
			fmt.Sprintf("The target file '%s' already exists.\n\n"+
				"Are you sure that you want to overwrite the existing file?",
				targetFileName)) {
			return "", fmt.Errorf("target file '%s' already exists", targetFileName)
		}
	}

	targetFilePath, err := r.repo.DumpFile(snapshot, file, targetPath)
	if err != nil {
		return "", err
	}
	return targetFilePath, nil
}

func (r *ResticBrowserApp) DumpFileToTemp(snapshotID string, file *restic.File) (string, error) {
	snapshot := r.snapshots[snapshotID]
	if snapshot == nil {
		return "", fmt.Errorf("%s is not a valid snapshot ID", snapshotID)
	}
	targetPath, err := os.MkdirTemp(r.tempPath, "dump")
	if err != nil {
		return "", err
	}
	targetFilePath, err := r.repo.DumpFile(snapshot, file, targetPath)
	if err != nil {
		return "", err
	}
	return targetFilePath, nil
}
