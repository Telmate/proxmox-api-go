package test_data_qemu

import (
	"strings"

	"golang.org/x/crypto/ssh"
)

type Test_AuthorizedKey struct {
	PublicKey ssh.PublicKey
	Comment   string
	Options   []string
}

func parse(key string) ssh.PublicKey {
	encoded, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		panic(err)
	}
	return encoded
}

func PublicKey_Decoded_Output() []Test_AuthorizedKey {
	return publicKey_Decoded()
}

func PublicKey_Decoded_Input() []Test_AuthorizedKey {
	return publicKey_Decoded()
}

func publicKey_Decoded() []Test_AuthorizedKey {
	return []Test_AuthorizedKey{
		{PublicKey: parse("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT7lsC9gTAjL0FUPlHqnz71TzqDMdsdHhWu54M7NN4E9KNzKwzUy1h6ZuOMm+d0nWX+yuT2Mfzi8NaKe5ATg0bwmrzZ1ikS/tGs7v/TyMSBOlmrS5v0g8rn40bphCqnNeNcfP9JR2zyq4UccpdIYA62t6Ky9d/WBbsAQRESwZVhpU9JGhwnVHFcNN5svlDwz9wzW1a2J2/E76+vym+3Rt4W9s3MqQZdbHozo4N43puXq7PH1tTr/RT84uaMF4XLx1CUm+bMZLgtac8sHl1DJz4gC3MLasD6UXZzRz99K+QAHD6YsXHDwdWu6QAkqzS0DNDbm0E618wn4GEZAJJhehh"),
			Options: []string{`no-pty`, `command="/path/to/script.sh"`},
			Comment: "test@VScode"},
		{PublicKey: parse("ssh-dss AAAAB3NzaC1kc3MAAACBAN6VwM2CMPrpz0CT8z4UP5we4Jt1MSDHumArdzTaxaqtAcV6Z+a4ZO/0geqEDZJSideX7Iq8zYrzdXGXfR+8N5GHoz49mVFit101cKAvcwZhzVeXQ1Cc8Zyjk53qmjWiNonfsjxP9VorNjjb/zGnA3ZnazflfyzqwEr8fV7JtUwjAAAAFQDlk3FT+QmsKiiBjBuekwyFeVzwiwAAAIBeAlzP9hsVeEbPjEjkxi9/hVgNQE8xtuUMZUCq7NOu5RlGzPHStzh8ByMh0Jsly0GbVHUfM84ikSpU/L5O3j75vq+cng77mezAGWfHfBpAL+whKfXvYHy0mqb0M1krzbdRbQkt9TV4gNw+Nac17jmfnRBebYYoJltehCognAU+xAAAAIEAmI1SEcjqSTHRnHeypg08ppcpRUGx0Mkcb/Moos2SVfSfWBXrNR7p6eRzVPN0gCXSLsiaE0DaRvM+GPRJeffCh4+Ahx84Gptf0m+EXH47sPfsumk8XxItDZa4zYYJ2gAISBdLD06iMtmJWAzD59FXDaHedxom9/Hb7oQXHEUzQQY="),
			Comment: "test@VScode"},
		{PublicKey: parse("ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAF9dgZNa82njYtBR2zhCQs1yHL/GqA+AAmz97bjj2t2EQwMepx3TT8RubZscqwt6yedPREJU/8x0XtoEWkQzjBkGgCc2ip8xGyy6j3Th9YtYj9gW1g7Rwmqwnz0ZOd/l12tC3q7ujS7hlInkhxbOyhqNXZ+obseOaS0g5Toqpgr+mV1Rg==")},
		{PublicKey: parse("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"),
			Comment: "cardno:18 228 342"},
	}
}

func PublicKey_Encoded_Output() string {
	return strings.Join([]string{
		"no-pty%2Ccommand%3D%22%2Fpath%2Fto%2Fscript.sh%22%20ssh-rsa%20AAAAB3NzaC1yc2EAAAADAQABAAABAQDT7lsC9gTAjL0FUPlHqnz71TzqDMdsdHhWu54M7NN4E9KNzKwzUy1h6ZuOMm%2Bd0nWX%2ByuT2Mfzi8NaKe5ATg0bwmrzZ1ikS%2FtGs7v%2FTyMSBOlmrS5v0g8rn40bphCqnNeNcfP9JR2zyq4UccpdIYA62t6Ky9d%2FWBbsAQRESwZVhpU9JGhwnVHFcNN5svlDwz9wzW1a2J2%2FE76%2Bvym%2B3Rt4W9s3MqQZdbHozo4N43puXq7PH1tTr%2FRT84uaMF4XLx1CUm%2BbMZLgtac8sHl1DJz4gC3MLasD6UXZzRz99K%2BQAHD6YsXHDwdWu6QAkqzS0DNDbm0E618wn4GEZAJJhehh%20test%40VScode",
		"ssh-dss%20AAAAB3NzaC1kc3MAAACBAN6VwM2CMPrpz0CT8z4UP5we4Jt1MSDHumArdzTaxaqtAcV6Z%2Ba4ZO%2F0geqEDZJSideX7Iq8zYrzdXGXfR%2B8N5GHoz49mVFit101cKAvcwZhzVeXQ1Cc8Zyjk53qmjWiNonfsjxP9VorNjjb%2FzGnA3ZnazflfyzqwEr8fV7JtUwjAAAAFQDlk3FT%2BQmsKiiBjBuekwyFeVzwiwAAAIBeAlzP9hsVeEbPjEjkxi9%2FhVgNQE8xtuUMZUCq7NOu5RlGzPHStzh8ByMh0Jsly0GbVHUfM84ikSpU%2FL5O3j75vq%2Bcng77mezAGWfHfBpAL%2BwhKfXvYHy0mqb0M1krzbdRbQkt9TV4gNw%2BNac17jmfnRBebYYoJltehCognAU%2BxAAAAIEAmI1SEcjqSTHRnHeypg08ppcpRUGx0Mkcb%2FMoos2SVfSfWBXrNR7p6eRzVPN0gCXSLsiaE0DaRvM%2BGPRJeffCh4%2BAhx84Gptf0m%2BEXH47sPfsumk8XxItDZa4zYYJ2gAISBdLD06iMtmJWAzD59FXDaHedxom9%2FHb7oQXHEUzQQY%3D%20test%40VScode",
		"ecdsa-sha2-nistp521%20AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAF9dgZNa82njYtBR2zhCQs1yHL%2FGqA%2BAAmz97bjj2t2EQwMepx3TT8RubZscqwt6yedPREJU%2F8x0XtoEWkQzjBkGgCc2ip8xGyy6j3Th9YtYj9gW1g7Rwmqwnz0ZOd%2Fl12tC3q7ujS7hlInkhxbOyhqNXZ%2BobseOaS0g5Toqpgr%2BmV1Rg%3D%3D",
		"ssh-ed25519%20AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW%20cardno%3A18%20228%20342"},
		"%0A") + "%0A"
}

func PublicKey_Encoded_Input() string {
	return strings.Join([]string{
		"no-pty%2Ccommand%3D%22%2Fpath%2Fto%2Fscript.sh%22%20ssh-rsa%20AAAAB3NzaC1yc2EAAAADAQABAAABAQDT7lsC9gTAjL0FUPlHqnz71TzqDMdsdHhWu54M7NN4E9KNzKwzUy1h6ZuOMm%2Bd0nWX%2ByuT2Mfzi8NaKe5ATg0bwmrzZ1ikS%2FtGs7v%2FTyMSBOlmrS5v0g8rn40bphCqnNeNcfP9JR2zyq4UccpdIYA62t6Ky9d%2FWBbsAQRESwZVhpU9JGhwnVHFcNN5svlDwz9wzW1a2J2%2FE76%2Bvym%2B3Rt4W9s3MqQZdbHozo4N43puXq7PH1tTr%2FRT84uaMF4XLx1CUm%2BbMZLgtac8sHl1DJz4gC3MLasD6UXZzRz99K%2BQAHD6YsXHDwdWu6QAkqzS0DNDbm0E618wn4GEZAJJhehh%20test%40VScode",
		"ssh-dss%20%20%20%20AAAAB3NzaC1kc3MAAACBAN6VwM2CMPrpz0CT8z4UP5we4Jt1MSDHumArdzTaxaqtAcV6Z%2Ba4ZO%2F0geqEDZJSideX7Iq8zYrzdXGXfR%2B8N5GHoz49mVFit101cKAvcwZhzVeXQ1Cc8Zyjk53qmjWiNonfsjxP9VorNjjb%2FzGnA3ZnazflfyzqwEr8fV7JtUwjAAAAFQDlk3FT%2BQmsKiiBjBuekwyFeVzwiwAAAIBeAlzP9hsVeEbPjEjkxi9%2FhVgNQE8xtuUMZUCq7NOu5RlGzPHStzh8ByMh0Jsly0GbVHUfM84ikSpU%2FL5O3j75vq%2Bcng77mezAGWfHfBpAL%2BwhKfXvYHy0mqb0M1krzbdRbQkt9TV4gNw%2BNac17jmfnRBebYYoJltehCognAU%2BxAAAAIEAmI1SEcjqSTHRnHeypg08ppcpRUGx0Mkcb%2FMoos2SVfSfWBXrNR7p6eRzVPN0gCXSLsiaE0DaRvM%2BGPRJeffCh4%2BAhx84Gptf0m%2BEXH47sPfsumk8XxItDZa4zYYJ2gAISBdLD06iMtmJWAzD59FXDaHedxom9%2FHb7oQXHEUzQQY%3D%20test%40VScode",
		"ecdsa-sha2-nistp521%20AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAF9dgZNa82njYtBR2zhCQs1yHL%2FGqA%2BAAmz97bjj2t2EQwMepx3TT8RubZscqwt6yedPREJU%2F8x0XtoEWkQzjBkGgCc2ip8xGyy6j3Th9YtYj9gW1g7Rwmqwnz0ZOd%2Fl12tC3q7ujS7hlInkhxbOyhqNXZ%2BobseOaS0g5Toqpgr%2BmV1Rg%3D%3D",
		"%0A",
		"%0A",
		"ssh-ed25519%20AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW%20%20%20%20%20cardno%3A18%20228%20342"},
		"%0A") + "%0A"
}
