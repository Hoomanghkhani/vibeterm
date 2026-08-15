package ui

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/models"
)

// ShowHostDialog opens a modal for creating or editing an infrastructure node
func ShowHostDialog(parent fyne.Window, initialHost *models.Host, onSave func(models.Host)) {
	nameEntry := widget.NewEntry()
	hostnameEntry := widget.NewEntry()
	portEntry := widget.NewEntry()
	portEntry.SetText("22")
	userEntry := widget.NewEntry()
	userEntry.SetText("root")

	authSelect := widget.NewSelect([]string{"password", "private_key", "ssh_agent", "certificate", "hardware_key"}, nil)
	authSelect.SetSelected("password")

	passwordEntry := widget.NewPasswordEntry()
	keyPathEntry := widget.NewEntry()
	keyPassEntry := widget.NewPasswordEntry()

	envSelect := widget.NewSelect([]string{"production", "staging", "dev", "edge"}, nil)
	envSelect.SetSelected("production")

	folderEntry := widget.NewEntry()
	folderEntry.SetText("Default")

	tagsEntry := widget.NewEntry()
	tagsEntry.SetPlaceHolder("e.g. aws, k8s, web")

	jumpHostEntry := widget.NewEntry()
	jumpHostEntry.SetPlaceHolder("Bastion hostname (optional multi-hop)")

	if initialHost != nil {
		nameEntry.SetText(initialHost.Name)
		hostnameEntry.SetText(initialHost.Hostname)
		portEntry.SetText(strconv.Itoa(initialHost.Port))
		userEntry.SetText(initialHost.Username)
		authSelect.SetSelected(string(initialHost.AuthMethod))
		passwordEntry.SetText(initialHost.Password)
		keyPathEntry.SetText(initialHost.PrivateKeyPath)
		keyPassEntry.SetText(initialHost.KeyPassphrase)
		envSelect.SetSelected(initialHost.Environment)
		folderEntry.SetText(initialHost.Folder)
		tagsEntry.SetText(strings.Join(initialHost.Tags, ", "))
		if len(initialHost.JumpChain) > 0 {
			jumpHostEntry.SetText(initialHost.JumpChain[0].Hostname)
		}
	}

	form := widget.NewForm(
		widget.NewFormItem("Host Name", nameEntry),
		widget.NewFormItem("Hostname / IP", hostnameEntry),
		widget.NewFormItem("Port", portEntry),
		widget.NewFormItem("Username", userEntry),
		widget.NewFormItem("Auth Method", authSelect),
		widget.NewFormItem("Password", passwordEntry),
		widget.NewFormItem("Private Key Path", keyPathEntry),
		widget.NewFormItem("Passphrase", keyPassEntry),
		widget.NewFormItem("Jump Host / Bastion", jumpHostEntry),
		widget.NewFormItem("Environment", envSelect),
		widget.NewFormItem("Folder / Group", folderEntry),
		widget.NewFormItem("Tags (comma sep)", tagsEntry),
	)

	d := dialog.NewCustomConfirm("Configure Host", "Save Host", "Cancel", container.NewVScroll(form), func(confirmed bool) {
		if !confirmed {
			return
		}

		port, _ := strconv.Atoi(portEntry.Text)
		if port <= 0 {
			port = 22
		}

		var tags []string
		for _, t := range strings.Split(tagsEntry.Text, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}

		host := models.Host{
			Name:           nameEntry.Text,
			Hostname:       hostnameEntry.Text,
			Port:           port,
			Protocol:       models.ProtocolSSH,
			Username:       userEntry.Text,
			AuthMethod:     models.AuthMethod(authSelect.Selected),
			Password:       passwordEntry.Text,
			PrivateKeyPath: keyPathEntry.Text,
			KeyPassphrase:  keyPassEntry.Text,
			Environment:    envSelect.Selected,
			Folder:         folderEntry.Text,
			Tags:           tags,
			Color:          "#00F0FF",
			Health:         models.HealthOnline,
		}

		if initialHost != nil && initialHost.ID != "" {
			host.ID = initialHost.ID
		}

		if jh := strings.TrimSpace(jumpHostEntry.Text); jh != "" {
			host.JumpChain = []models.JumpHostHop{
				{
					HopIndex:   0,
					Hostname:   jh,
					Port:       22,
					Username:   host.Username,
					AuthMethod: host.AuthMethod,
				},
			}
		}

		if onSave != nil {
			onSave(host)
		}
	}, parent)

	d.Resize(fyne.NewSize(520, 520))
	d.Show()
}
