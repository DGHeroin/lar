package larc

import "github.com/DGHeroin/golua/lua"

type Mail struct {
    L *lua.State
}

func (m Mail) pop(L *lua.State) int {
    return 0
}

func (m Mail) push(L *lua.State) int {
    return 0
}

func newMail(L *lua.State) *Mail {
    m := &Mail{
        L: L,
    }
    L.PushGoFunction(m.pop)
    L.SetGlobal("mail_pop")
    L.PushGoFunction(m.push)
    L.SetGlobal("mail_push")
    return m
}
