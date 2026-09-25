//go:build race

package bookmaker

// raceSlowdown — bộ dò data race (-race) làm code chạy chậm cỡ 5–10 lần;
// test đo thời gian nhân ngưỡng theo hệ số này.
const raceSlowdown = 10
