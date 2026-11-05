// pwreset —— 用户管理 CLI：list / setpw / verify / rm / promote
package main

import (
	"flag"
	"fmt"
	"os"

	"nova/internal/auth"
)

func main() {
	dir := flag.String("dir", "data", "数据目录（含 users.json）")
	flag.Parse()
	cmd := flag.Arg(0)
	st, err := auth.OpenStore(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch cmd {
	case "list":
		for _, u := range st.List() {
			fmt.Printf("%s\t%s\t%s\tdisabled=%v\n", u.Email, u.Role, u.CreatedAt, u.Disabled)
		}
	case "setpw":
		email, pw := flag.Arg(1), flag.Arg(2)
		if email == "" || pw == "" {
			fmt.Fprintln(os.Stderr, "usage: pwreset setpw <email> <newpw>")
			os.Exit(1)
		}
		if err := st.SetPassword(email, pw); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("OK password updated:", email)
	case "verify":
		email, pw := flag.Arg(1), flag.Arg(2)
		u, ok := st.ByEmail(email)
		if !ok || !auth.VerifyPassword(u.Hash, pw) {
			fmt.Println("MISMATCH")
			os.Exit(1)
		}
		fmt.Println("OK password verified:", email)
	case "rm":
		email := flag.Arg(1)
		if !st.Remove(email) {
			fmt.Fprintln(os.Stderr, "not found:", email)
			os.Exit(1)
		}
		fmt.Println("OK removed:", email)
	case "promote":
		email := flag.Arg(1)
		if err := st.SetRole(email, "super"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("OK promoted to super:", email)
	default:
		fmt.Fprintln(os.Stderr, "usage: pwreset list|setpw <email> <pw>|verify <email> <pw>|rm <email>|promote <email>")
		os.Exit(1)
	}
}
