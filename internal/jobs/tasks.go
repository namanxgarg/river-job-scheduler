package jobs

import (
    "errors"
    "fmt"
    "math/rand"
    "strings"
    "time"
)

func runExampleTask(payload string) (string, error) {
    // 50% chance to fail for demonstration
    if rand.Float32() < 0.5 {
        return "", errors.New("random failure, try again")
    }
    time.Sleep(2 * time.Second)
    result := strings.ToUpper(payload)
    return fmt.Sprintf("Processed: %s", result), nil
}
