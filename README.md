# SAGO - opinionated implementation of the saga pattern in go

## What it does

It allows you to run a series of actions as long as you have undo actions for them. If one of the actions you perform fails, the compensating actions for that action will be performed and then step by step backward for all previous steps in reverse order.

## Why opinionated?

Because the compensation steps will use the results of the action step as the arguments.

## Example usage

```go

func stringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func floatToString(i float64) (string, error) {
	return fmt.Sprintf("%d", int(i)), nil
}

func incr(i int) (int, error) {
	return int(i) + 1, nil
}

func decr(i float64) (int, error) {
	return int(i) - 1, nil
}


var log = sago.Log{
	ReadWriter: &bytes.Buffer{},
	LogItems:   []LogItem{},
}

sec = sago.NewSEC("test", log)
sec.AddAction("atoi", stringToInt, floatToString)
sec.AddAction("incr", incr, decr)

```

## Gotchas

Because of the lack of generics, the package makes havy usage of reflection. 
One problem that arises from this is that when an interface is reverted to a number, it will always be reverted to a ```float64```.
This is the reason you see in the example the function ```floatToString``` instead of seeing ```intToString```.
You must take care of this generalization that happens in reflection when you design your compensation function.