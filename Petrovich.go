// МОДИФИЦИРОВАННАЯ ВЕРСИЯ ПАКЕТА

package Petrovich

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Rules struct {
	Lastname   RulesGroup `json:"lastname"`
	Firstname  RulesGroup `json:"firstname"`
	Middlename RulesGroup `json:"middlename"`
}

type RulesGroup struct {
	Exceptions []Rule `json:"exceptions"`
	Suffixes   []Rule `json:"suffixes"`
}

type Rule struct {
	Gender string   `json:"gender"`
	Test   []string `json:"test"`
	Mods   []string `json:"mods"`
	Tags   []string `json:"tags"`
}

// падежи инициализируются константами от 0 до 4
const (
	// Родительный
	Genitive = iota
	// Дательный
	Dative
	// Винительный
	Accusative
	// Творительный
	Instrumental
	// Предложный
	Prepositional
)

func LoadRules(FileWithRules string) (*Rules, error) {

	rulesFile, err := os.Open(FileWithRules)
	if err != nil {
		return nil, err
	}
	rulesData, _ := ioutil.ReadAll(rulesFile)
	if err != nil {
		return nil, err
	}
	defer rulesFile.Close()

	var r Rules

	err = json.Unmarshal([]byte(rulesData), &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

/*
Склонение имени
value: Значение для склонения
gCase: Падеж для склонения
gender: Грамматический род
*/
func (r *Rules) InfFirstname(value string, gCase int, gender string) string {

	return inflect(value, r.Firstname, gCase, gender)
}

/*
Склонение фамилии
value: Значение для склонения
gCase: Падеж для склонения
gender: Грамматический род
*/
func (r *Rules) InfLastname(value string, gCase int, gender string) string {

	return inflect(value, r.Lastname, gCase, gender)
}

/*
Склонение отчества
value: Значение для склонения
gCase: Падеж для склонения
gender: Грамматический род
*/
func (r *Rules) InfMiddlename(value string, gCase int, gender string) string {

	return inflect(value, r.Middlename, gCase, gender)
}

/*
Склонение ФИО
value: ФИО через проблеы
gCase: Падеж для склонения
short: Результат в сокращенной форме "Иванов И.И."
*/
func (r *Rules) InfFio(fio string, gCase int, short bool) (string, error) {
	result := ""
	fio = strings.Trim(fio, " ")

	fioArray := strings.Split(fio, " ")
	if len(fioArray) != 3 {
		if gCase == 0 {
			fmt.Println("Error format of fio [Lastname FirstName MiddleName], trying InfFi")
		}
		return result, errors.New("Error format of fio [Lastname FirstName MiddleName]")
	}

	gender := detectGender(fioArray[2])

	fioArray[0] = inflect(fioArray[0], r.Lastname, gCase, gender)

	if short {
		result = fmt.Sprintf("%s %s.%s.", fioArray[0], string([]rune(fioArray[1])[:1]), string([]rune(fioArray[2])[:1]))
	} else {
		fioArray[1] = inflect(fioArray[1], r.Firstname, gCase, gender)
		fioArray[2] = inflect(fioArray[2], r.Middlename, gCase, gender)
		result = strings.Join(fioArray, " ")
	}
	return result, nil
}

func (r *Rules) InfFi(fi string, gCase int, short bool) string {
	result := ""
	fi = strings.Trim(fi, " ")

	fioArray := strings.Split(fi, " ")

	fioArray = deleteEmpty(fioArray)
	if len(fioArray) != 2 {
		fmt.Println("Error format of fio [Lastname FirstName]")
		return result
	}

	gender := "androgynous"

	fioArray[0] = inflect(fioArray[0], r.Lastname, gCase, gender)

	if short {
		result = fmt.Sprintf("%s %s.%s.", fioArray[0], string([]rune(fioArray[1])[:1]))
	} else {
		fioArray[1] = inflect(fioArray[1], r.Firstname, gCase, gender)
		result = strings.Join(fioArray, " ")
	}
	return result
}

func detectGender(middlename string) string {
	if strings.HasSuffix(middlename, "ич") {
		return "male"
	}
	if strings.HasSuffix(middlename, "на") {
		return "female"
	}
	return "androgynous"
}

func inflect(value string, nameFormRules RulesGroup, gCase int, gender string) string {

	if result := checkExcludes(value, nameFormRules, gCase, gender); result != "" {
		return result
	}

	value = strings.Trim(value, " ")
	parts := strings.Split(value, "-")

	if len(parts) > 1 {
		for i := 0; i < len(parts); i++ {
			parts[i] = findRules(parts[i], nameFormRules, gCase, gender)
		}
		return strings.Join(parts, "-")
	}
	return findRules(value, nameFormRules, gCase, gender)
}

func checkExcludes(name string, rGroup RulesGroup, gCase int, gender string) string {
	lowerName := strings.ToLower(name)

	for _, exception := range rGroup.Exceptions {
		if exception.Gender == gender || exception.Gender == "androgynous" {
			for _, t := range exception.Test {
				if t == lowerName {
					return applyRule(exception.Mods[gCase], name)
				}
			}
		}
	}
	return ""
}

func applyRule(mod string, name string) string {
	if mod == "." {
		return name
	}
	runnedName := []rune(name)
	runnedName = runnedName[:len(runnedName)-strings.Count(mod, "-")]
	result := string(runnedName) + strings.Replace(mod, "-", "", -1)
	return result
}

func findRules(name string, rGroup RulesGroup, gCase int, gender string) string {

	for _, suffix := range rGroup.Suffixes {
		if gender == suffix.Gender || suffix.Gender == "androgynous" {
			for _, str := range suffix.Test {
				if len(str) < len(name) {

					lastChar := name[len(name)-len(str):]

					if lastChar == str {
						if suffix.Mods[gCase] == "." {
							continue
						}
						return applyRule(suffix.Mods[gCase], name)
					}
				}
			}
		}
	}
	return name
}

func deleteEmpty(src []string) []string {
	var dst []string
	for _, str := range src {
		if str != "" {
			dst = append(dst, str)
		}
	}
	return dst
}
