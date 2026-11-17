package ui

type language string

const (
	english language = "en"
	russian language = "ru"
)

var translations = map[language]map[string]string{
	english: {
		"resize": "Resize the terminal to at least 72 columns by 16 rows.", "search": "Search", "commands": "Commands", "details": "Command Details", "tags": "Tags", "description": "Description", "command": "Command", "variables": "Variables", "none": "None", "danger": "Dangerous command", "no_results": "No commands match this search.", "more_above": "↑ more results above", "more_below": "↓ more results below",
		"footer":     "[Enter] Copy  [Ctrl+R] Run  [Tab] Tabs  [Ctrl+A] New  [Ctrl+E] Edit  [Ctrl+H] History  [Ctrl+L] Language  [Ctrl+O] About  [?] Help  [Esc] Quit",
		"help_title": "Keyboard Shortcuts", "navigate": "Navigate commands", "type_search": "Type     Search", "copy": "Copy command", "run": "Run command", "tabs": "Switch category", "new": "Create command", "edit": "Edit command", "history_key": "Import shell history", "language_key": "Switch language", "about_key": "About Snipbox", "close": "Close dialog / quit", "press_close": "Press Esc to close",
		"history_title": "Shell History — select a command", "history_footer": "Enter imports  ·  Esc closes", "new_title": "New Command", "edit_title": "Edit Command", "category": "Category", "name": "Name", "tags_form": "Tags (comma separated)", "tab_advances": "Tab advances  ·  Ctrl+S saves  ·  Esc cancels", "fill_variables": "Fill Variables", "enter_advances": "Enter advances  ·  Esc cancels", "confirm_title": "Confirm Execution", "danger_explainer": "This command can make destructive changes.", "review_command": "Review it before continuing:", "confirm_keys": "Press Y to run  ·  N / Esc to cancel",
		"about_title": "About Snipbox", "about_line1": "A thoughtfully crafted command library for your terminal.", "about_line2": "Search, understand, and run your everyday commands.", "author": "Created by Pavel Slepcov", "stack": "Built with Rust and Go", "version": "Cross-platform command workspace", "language_name": "Language",
		"copied": "✓ Copied to clipboard!", "saved": "✓ Command saved", "imported": "✓ Imported command from shell history", "completed": "✓ Command completed", "canceled": "Command canceled", "no_history": "No shell history found", "required": "Name and command are required", "clipboard_error": "Clipboard error: ", "save_error": "Could not save command: ", "usage_error": "Could not save usage count: ", "parse_error": "Could not parse command template: ", "render_error": "Could not render command: ", "decode_error": "Could not decode rendered command", "safety_error": "Could not check command safety: ", "failed": "Command failed: ",
	},
	russian: {
		"resize": "Увеличьте окно терминала: минимум 72 столбца и 16 строк.", "search": "Поиск", "commands": "Команды", "details": "Подробнее о команде", "tags": "Теги", "description": "Описание", "command": "Команда", "variables": "Переменные", "none": "Нет", "danger": "Опасная команда", "no_results": "По этому запросу команды не найдены.", "more_above": "↑ выше есть ещё команды", "more_below": "↓ ниже есть ещё команды",
		"footer":     "[Enter] Копировать  [Ctrl+R] Запуск  [Tab] Вкладки  [Ctrl+A] Создать  [Ctrl+E] Изменить  [Ctrl+H] История  [Ctrl+L] Язык  [Ctrl+O] О программе  [?] Справка  [Esc] Выход",
		"help_title": "Горячие клавиши", "navigate": "Перемещение по командам", "type_search": "Ввод     Поиск", "copy": "Копировать команду", "run": "Заполнить и выполнить", "tabs": "Переключить категорию", "new": "Создать команду", "edit": "Изменить команду", "history_key": "Импортировать историю команд", "language_key": "Переключить язык", "about_key": "О программе Snipbox", "close": "Закрыть окно / выйти", "press_close": "Нажмите Esc, чтобы закрыть",
		"history_title": "История команд — выберите команду", "history_footer": "Enter — импортировать  ·  Esc — закрыть", "new_title": "Новая команда", "edit_title": "Изменить команду", "category": "Категория", "name": "Название", "tags_form": "Теги (через запятую)", "tab_advances": "Tab — дальше  ·  Ctrl+S — сохранить  ·  Esc — отмена", "fill_variables": "Заполнить переменные", "enter_advances": "Enter — дальше  ·  Esc — отмена", "confirm_title": "Подтверждение запуска", "danger_explainer": "Команда может привести к удалению или потере данных.", "review_command": "Проверьте команду перед запуском:", "confirm_keys": "Y — выполнить  ·  N / Esc — отмена",
		"about_title": "О программе Snipbox", "about_line1": "Продуманная библиотека команд для вашего терминала.", "about_line2": "Ищите, изучайте и запускайте команды на каждый день.", "author": "Автор: Pavel Slepcov", "stack": "Создано на Rust и Go", "version": "Кроссплатформенная консольная утилита", "language_name": "Язык",
		"copied": "✓ Команда скопирована", "saved": "✓ Команда сохранена", "imported": "✓ Команда импортирована из истории", "completed": "✓ Команда выполнена", "canceled": "Выполнение отменено", "no_history": "История команд не найдена", "required": "Укажите название и команду", "clipboard_error": "Ошибка буфера обмена: ", "save_error": "Не удалось сохранить команду: ", "usage_error": "Не удалось сохранить счётчик запусков: ", "parse_error": "Не удалось разобрать шаблон команды: ", "render_error": "Не удалось подставить значения: ", "decode_error": "Не удалось прочитать готовую команду", "safety_error": "Не удалось проверить безопасность команды: ", "failed": "Команда завершилась с ошибкой: ",
	},
}

func tr(lang language, key string) string {
	if words, ok := translations[lang]; ok {
		if value, ok := words[key]; ok {
			return value
		}
	}
	return translations[english][key]
}
