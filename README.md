# SAGO - opinionated implementation of the saga pattern in go

## Important

Please use functions that return an error as the last parameter. This is the way ```sago``` checks for errors.

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

var sagaLog = sago.Log{
    ReadWriter: &bytes.Buffer{},
    LogItems:   []LogItem{},
}

 sec := sago.NewSEC("test", sagaLog)

func RunIterative(){

    sec.AddAction("atoi", stringToInt, floatToString)
    sec.AddAction("incr", incr, decr)

    rez, err := sec.Step("atoi", []interface{}{"10"})

    if err != nil {
        log.Println("Compensated from step 1")
        return
    }

    rez, err = sec.Step("incr", rez[:1])

    if err != nil {
        log.Println("Compensated from step 2")
        return
    }

    fmt.Println(sec.Result())

}

func RunSimple(){
    
	err := sec.AddAction("atoi", stringToInt, floatToString)
	err = sec.AddAction("incr", incr, decr)

	err = sec.
		Run("atoi", SagaStep{stringToInt, intToString}, []interface{}{"10"}).
		Run("incr", SagaStep{incr, decr}).
		Run("fail", SagaStep{stringToInt, intToString}). //this should fail as stringToInt will be fed the result of incr
        End()
    
    if err != nil {
        log.Println("Failed and not completely compensated")
        return
    } else if sec.IsCompensated(){
        log.Println("Failed and compensated")
    } else {
        fmt.Println("Completed successfully")
        fmt.Println(sec.Result())
    }
}

```

## Gotchas

Because of the lack of generics, the package makes havy usage of reflection. 
One problem that arises from this is that when an json number is reverted to a number, it will always be reverted to a ```float64```.
This is the reason you see in the example the function ```floatToString``` instead of seeing ```intToString``` (in the ```TestCompensateOK``` function).

You must take care of this generalization that happens in decoding a json when you design your compensation function. 

That is because normaly, your Log should write into a persistent location, not in the memory like we have here. And the log could be latter read from there and converted back into log items.