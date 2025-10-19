package screen

import (
	"context"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/component"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/util"
)

var (
	itemColHeight = 30
	itemColWidth  = util.ScreenWidth - 8
)

type InventoryScreen struct {
	keymap              util.KeyMap
	CharacterRepository repository.CharacterRepository
	Context             context.Context

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	characterId uuid.UUID
	copper      *component.SimpleIntComponent
	silver      *component.SimpleIntComponent
	electrum    *component.SimpleIntComponent
	gold        *component.SimpleIntComponent
	platinum    *component.SimpleIntComponent
	itemList    *list.List
}

func NewInventoryScreen(k util.KeyMap, cr repository.CharacterRepository, ctx context.Context, characterId uuid.UUID) *InventoryScreen {
	s := InventoryScreen{
		keymap:              k,
		CharacterRepository: cr,
		Context:             ctx,
		characterId:         characterId,
	}

	s.copper = component.NewSimpleIntComponent(k, "CP", 0, s.persistWalletField("copper"), true, true)
	s.silver = component.NewSimpleIntComponent(k, "SP", 0, s.persistWalletField("silver"), true, true)
	s.electrum = component.NewSimpleIntComponent(k, "EP", 0, s.persistWalletField("electrum"), true, true)
	s.gold = component.NewSimpleIntComponent(k, "GP", 0, s.persistWalletField("gold"), true, true)
	s.platinum = component.NewSimpleIntComponent(k, "PP", 0, s.persistWalletField("platinum"), true, true)
	return &s
}

func (s *InventoryScreen) Init() tea.Cmd {
	s.focusOn(s.copper)
	s.lastFocusedElement = s.copper

	return s.reloadDataCmd()
}

func (s *InventoryScreen) reloadDataCmd() tea.Cmd {
	return tea.Batch(command.LoadItemsCmd(s.CharacterRepository, s.Context, s.characterId),
		command.LoadWalletCommand(s.CharacterRepository, s.Context, s.characterId))
}

func (s *InventoryScreen) populateItems(items []models.ItemTO) {
	if s.itemList == nil {
		s.itemList = list.NewList(s.keymap,
			list.LeftAlignedListStyle).
			WithFixedWidth(itemColWidth).
			WithViewport(itemColHeight - 2)
	}
	s.itemList.WithRows(s.GetItemRows(items))
}

func (s *InventoryScreen) populateWallet(wallet models.WalletTO) {
	s.copper = component.NewSimpleIntComponent(s.keymap, "CP", wallet.Copper, s.persistWalletField("copper"), true, true)
	s.silver = component.NewSimpleIntComponent(s.keymap, "SP", wallet.Silver, s.persistWalletField("silver"), true, true)
	s.electrum = component.NewSimpleIntComponent(s.keymap, "EP", wallet.Electrum, s.persistWalletField("electrum"), true, true)
	s.gold = component.NewSimpleIntComponent(s.keymap, "GP", wallet.Gold, s.persistWalletField("gold"), true, true)
	s.platinum = component.NewSimpleIntComponent(s.keymap, "PP", wallet.Platinum, s.persistWalletField("platinum"), true, true)
}

func (s *InventoryScreen) persistWalletField(field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateWalletFields(s.Context, s.characterId, map[string]interface{}{field: v})
	}
}

func (s *InventoryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.DataOpMsg:
		if msg.Op != command.DataSave {
			cmd = s.reloadDataCmd()
		}
	case command.LoadWalletMsg:
		s.populateWallet(msg.Wallet)
	case command.LoadItemMsg:
		s.populateItems(msg.Items)
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "item") {
			cmd = command.DataOperationCommand(func() error {
				_, err := s.CharacterRepository.AddItem(s.Context, s.characterId, models.ItemTO{})
				return err
			}, command.DataCreate)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case editor.EditValueMsg:
		cmd = editor.SwitchToEditorCmd(msg.Editors)
	case tea.KeyMsg:
		switch s.focusedElement.(type) {
		case *list.List:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		default:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			case key.Matches(msg, s.keymap.Up):
				cmd = s.moveFocus(command.UpDirection)
			case key.Matches(msg, s.keymap.Down):
				cmd = s.moveFocus(command.DownDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		}
	}
	return s, cmd
}

func (s *InventoryScreen) View() string {
	topbar := s.RenderInventoryScreenTopBar()
	renderedItems := s.itemList.View()

	content := util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		Height(spellColHeight).
		Render(renderedItems)
	return lipgloss.JoinVertical(lipgloss.Left, topbar, content)
}

func (s *InventoryScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *InventoryScreen) moveFocus(d command.Direction) tea.Cmd {
	var cmd tea.Cmd
	s.Blur()

	switch s.lastFocusedElement {
	case s.copper:
		switch d {
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		case command.RightDirection:
			s.focusOn(s.silver)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.copper)
		}
	case s.silver:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.copper)
		case command.RightDirection:
			s.focusOn(s.electrum)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.silver)
		}
	case s.electrum:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.silver)
		case command.RightDirection:
			s.focusOn(s.gold)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.electrum)
		}
	case s.gold:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.electrum)
		case command.RightDirection:
			s.focusOn(s.platinum)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.gold)
		}
	case s.platinum:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.gold)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.platinum)
		}
	case s.itemList:
		switch d {
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		case command.UpDirection:
			s.focusOn(s.electrum)
		default:
			s.focusOn(s.itemList)
		}
	}
	return cmd
}

func (s *InventoryScreen) Focus() {
	s.focusOn(s.lastFocusedElement)
}

func (s *InventoryScreen) Blur() {
	if s.focusedElement != nil {
		s.focusedElement.Blur()
		s.lastFocusedElement = s.focusedElement
	}

	s.focusedElement = nil
}

func (s *InventoryScreen) GetItemRows(items []models.ItemTO) []list.Row {
	rows := []list.Row{}
	for i := range items {
		item := items[i]
		rows = append(rows, list.NewStructRow(s.keymap, item,
			RenderItemInfoRow,
			s.CreateItemEditors(item),
		).WithDestructor(DeleteItemCallback(s, item)).
			WithReader(RenderFullItemInfo))
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, "item"))
	return rows
}

func DeleteItemCallback(s *InventoryScreen, i models.ItemTO) func() tea.Cmd {
	return func() tea.Cmd {
		return command.DataOperationCommand(func() error { return s.CharacterRepository.DeleteItem(s.Context, i.ID) }, command.DataDelete)
	}
}

func (s *InventoryScreen) CreateItemEditors(item models.ItemTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(s.keymap, "Name", item.Name, s.persistItemStringField(item.ID, "name")),
		editor.NewEnumEditor(s.keymap, EquippedSymbols, "Equipped", int(item.Equipped), s.persistItemIntField(item.ID, "equipped")),
		editor.NewEnumEditor(s.keymap, AttunementSymbols, "Attunement Slots", item.AttunementSlots, s.persistItemIntField(item.ID, "attunement_slots")),
		editor.NewIntEditor(s.keymap, "Quantity", item.Quantity, s.persistItemIntField(item.ID, "quantity")),
		editor.NewStringEditor(s.keymap, "Description", item.Description, s.persistItemStringField(item.ID, "description")),
	}
}

func (s *InventoryScreen) persistItemStringField(id uuid.UUID, field string) func(string) error {
	return func(v string) error {
		return s.CharacterRepository.UpdateItemFields(s.Context, id, map[string]interface{}{field: v})
	}
}

func (s *InventoryScreen) persistItemIntField(id uuid.UUID, field string) func(int) error {
	return func(v int) error {
		return s.CharacterRepository.UpdateItemFields(s.Context, id, map[string]interface{}{field: v})
	}
}

func (s *InventoryScreen) RenderInventoryScreenTopBar() string {
	separator := util.GrayTextStyle.Width(6).Render(util.MakeVerticalSeparator(1))
	return util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		AlignHorizontal(lipgloss.Left).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Center,
			util.ForceWidth(s.copper.View(), 15),
			separator,
			util.ForceWidth(s.silver.View(), 15),
			separator,
			util.ForceWidth(s.electrum.View(), 15),
			separator,
			util.ForceWidth(s.gold.View(), 15),
			separator,
			util.ForceWidth(s.platinum.View(), 15),
		))
}

func RenderItemInfoRow(i models.ItemTO) string {
	values := []string{DrawItemPrefix(i), i.Name, DrawAttunementSlots(i.AttunementSlots)}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return strings.Join(values, " ∙ ")
}

func RenderFullItemInfo(i models.ItemTO) string {
	separator := util.MakeHorizontalSeparator(util.SmallScreenWidth-4, 1)
	content := strings.Join(
		[]string{
			i.Name,
			separator,
			"Equipped: " + DrawItemPrefix(i),
			separator,
			"Attunement slots required: " + DrawAttunementSlots(i.AttunementSlots),
			separator,
			"Quantity: " + strconv.Itoa(i.Quantity),
			separator,
			i.Description,
		},
		"\n")
	return util.DefaultTextStyle.
		AlignHorizontal(lipgloss.Left).
		Render(content)
}

func DrawItemPrefix(i models.ItemTO) string {
	s := ""
	switch i.Equipped {
	case models.Equipped:
		s = "■"
	case models.NotEquipped:
		s = "□"
	default:
		s = strconv.Itoa(i.Quantity)

	}
	return s
}

var EquippedSymbols []editor.EnumMapping = []editor.EnumMapping{
	{Value: int(models.NonEquippable), Label: "Not Equippable"},
	{Value: int(models.NotEquipped), Label: "Not Equipped"},
	{Value: int(models.Equipped), Label: "Equipped"},
}

var AttunementSymbols []editor.EnumMapping = []editor.EnumMapping{
	{Value: 0, Label: "□□□"},
	{Value: 1, Label: "■□□"},
	{Value: 2, Label: "■■□"},
	{Value: 3, Label: "■■■"},
}

func DrawAttunementSlots(used int) string {
	if used == 0 {
		return ""
	} else {
		s := strings.Repeat("■", used)
		s += strings.Repeat("□", 3-used)
		return s
	}
}
