package activitypublog

import "github.com/labstack/echo/v4"

// クライアントが非ログインならログインページにリダイレクトする
// ログイン済みならtokenとhostを返す
func RequireLoggedIn(c echo.Context) (string, string, error) {
	tokenCookie, err := c.Cookie("token")
	if err != nil {
		return "", "", c.Redirect(302, "/login")
	}
	token := tokenCookie.Value
	hostCookie, err := c.Cookie("host")
	if err != nil {
		return "", "", c.Redirect(302, "/login")
	}
	host := hostCookie.Value
	return token, host, nil
}
