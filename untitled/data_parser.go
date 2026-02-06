package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MonthlyRecord struct {
	Year   int
	Month  int
	Values map[string]float64
}

type MonthlyData struct {
	Header  []string
	Records []MonthlyRecord
}

type TrendData struct {
	Header []string
	Values map[string]float64
}

func ReadTemperatureFile(path string) (MonthlyData, TrendData, error) {
	file, err := os.Open(path)
	if err != nil {
		return MonthlyData{}, TrendData{}, fmt.Errorf("open temperature file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var monthlyHeader []string
	var monthlyNames []string
	var trendHeader []string
	var trendNames []string
	var records []MonthlyRecord
	var trendValues map[string]float64

	foundMonthlyHeader := false
	foundTrendHeader := false
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if !foundMonthlyHeader {
			if fields[0] == "Year" && len(fields) > 2 {
				monthlyHeader = fields
				fmt.Println(monthlyHeader) // testing
				monthlyNames = buildColumnNames(fields)
				foundMonthlyHeader = true
			}
			continue
		}

		if fields[0] == "Year" && len(fields) > 2 {
			trendHeader = fields
			fmt.Println(trendHeader) // testing
			trendNames = buildColumnNames(fields)
			foundTrendHeader = true
			continue
		}

		if fields[0] == "Trend" {
			if !foundTrendHeader {
				trendNames = monthlyNames
			}
			values, err := parseValues(fields[1:], trendNames, lineNo)
			if err != nil {
				return MonthlyData{}, TrendData{}, err
			}
			trendValues = values
			break
		}

		if foundTrendHeader {
			continue
		}

		if len(fields) < 2 {
			return MonthlyData{}, TrendData{}, fmt.Errorf("line %d: expected year and month", lineNo)
		}
		year, err := strconv.Atoi(fields[0])
		if err != nil {
			return MonthlyData{}, TrendData{}, fmt.Errorf("line %d: parse year: %w", lineNo, err)
		}
		month, err := strconv.Atoi(fields[1])
		if err != nil {
			return MonthlyData{}, TrendData{}, fmt.Errorf("line %d: parse month: %w", lineNo, err)
		}

		values, err := parseValues(fields[2:], monthlyNames, lineNo)
		if err != nil {
			return MonthlyData{}, TrendData{}, err
		}
		records = append(records, MonthlyRecord{
			Year:   year,
			Month:  month,
			Values: values,
		})
	}

	if err := scanner.Err(); err != nil {
		return MonthlyData{}, TrendData{}, fmt.Errorf("scan temperature file: %w", err)
	}

	if !foundMonthlyHeader {
		return MonthlyData{}, TrendData{}, fmt.Errorf("missing monthly header row")
	}
	if len(records) == 0 {
		return MonthlyData{}, TrendData{}, fmt.Errorf("no monthly records parsed")
	}
	if trendValues == nil {
		return MonthlyData{}, TrendData{}, fmt.Errorf("missing trend data row")
	}

	return MonthlyData{
			Header:  monthlyNames,
			Records: records,
		}, TrendData{
			Header: trendNames,
			Values: trendValues,
		}, nil
}

func buildColumnNames(headerFields []string) []string {
	if len(headerFields) <= 2 {
		return nil
	}
	valueFields := headerFields[2:]
	names := make([]string, 0, len(valueFields))

	currentRegion := ""
	for _, field := range valueFields {
		switch field {
		case "Land", "Ocean":
			if currentRegion == "" {
				names = append(names, field)
			} else {
				names = append(names, currentRegion+field)
			}
		default:
			currentRegion = field
			names = append(names, field)
		}
	}

	return names
}

func parseValues(fields []string, names []string, lineNo int) (map[string]float64, error) {
	if len(fields) != len(names) {
		return nil, fmt.Errorf("line %d: expected %d values, got %d", lineNo, len(names), len(fields))
	}

	values := make(map[string]float64, len(names))
	for i, name := range names {
		value, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: parse %s: %w", lineNo, name, err)
		}
		values[name] = value
	}
	return values, nil
}
