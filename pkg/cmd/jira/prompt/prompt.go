package prompt

import "github.com/AlecAivazis/survey/v2"

const (
	yes = "yes"
	no  = "no"
)

type Prompter interface {
	Select(string, []string) (string, error)
	SelectWithHelp(string, string, []string) (string, error)
	Input(string, string, ...survey.AskOpt) (string, error)
	InputWithHelp(string, string, string, ...survey.AskOpt) (string, error)
	Confirm(string) (bool, error)
}

func NewPrompter() Prompter {
	return &surveyPrompter{}
}

// prompter implementation
type surveyPrompter struct{}

func (p *surveyPrompter) Select(message string, options []string) (result string, err error) {
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}

	err = askSurvey(prompt, &result)
	return
}

func (p *surveyPrompter) SelectWithHelp(message string, help string, options []string) (result string, err error) {
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Help:    help,
	}

	err = askSurvey(prompt, &result)
	return
}

func (p *surveyPrompter) Input(message string, defaultValue string, ops ...survey.AskOpt) (result string, err error) {
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}

	err = askSurvey(prompt, &result, ops...)
	return
}

func (p *surveyPrompter) InputWithHelp(message, defaultValue, help string, ops ...survey.AskOpt) (result string, err error) {
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
		Help:    help,
	}

	err = askSurvey(prompt, &result, ops...)
	return
}

func (p *surveyPrompter) Confirm(message string) (bool, error) {
	choice, err := p.Select(message, []string{yes, no})
	if err != nil {
		return false, err
	}

	if choice == yes {
		return true, nil
	}

	return false, nil
}

func askSurvey(p survey.Prompt, r interface{}, ops ...survey.AskOpt) error {
	ops = append(ops, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = ">>"
		icons.Question.Format = "green+hb"
	}))
	return survey.AskOne(p, r, ops...)
}
