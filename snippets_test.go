package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSnippets(t *testing.T) {
	s := `We are immersed in a sea of dead people. All the dead that have gone before us, silent now, just staring, gaping. As we move and talk and fret, never once stopping to ask ourselves – or them! – what it was all about. Instead we drown ourselves in noise. Incessantly we babble, surrounded by false friends claiming that all is well. And look at us! Yes, we are well. Patting our backs and expecting a pat – and we do! – we smugly do enjoy.`

	h := `We are immersed in a sea of dead people. <b>All</b> the dead that have gone before us, silent now, just … to ask ourselves – or them! – what it was <b>all</b> about. Instead we drown ourselves in no<b>is</b>e. … surrounded by false friends claiming that <b>all</b> <b>is</b> <b>well</b>. And look at us! Yes, we are <b>well</b>. …`

	q := "all is well"
	r := snippets(q, s)
	assert.Equal(t, h, r)
}

func TestSnippetsLong(t *testing.T) {
	s := `VWwXetig mty8fORN UNia4NFm SQsfyFHk BLDdgVnc AcvKP2fs q8KxPH1A IaCzFj96 J0S2fqca jp3ElV9f ULIZ1aMX GSUO1pLI p7vuJie8 kPfc0ONq EthfIUjm u74guCZ8 IiJYxlR6 5j5LlapY TGO98fOQ fO2RUb1g W8zaPa0v ps0haNzW OOeFwf1h 1N3td7zk 0OoMX8Ek aTd3Ciea 2T1aK9WH QbYfUojs nP59gqvR tqoEK3vJ zJ7JmRby qKReayLo 9BIwFgID 4Q4Tk3HH 1VLdDzSx q0hKUOKm vWkUXz9S 684uXanc gIaJNRFc gabtBO9A EhIh4VtT gJ3p9LYL jPVFqc65 QmMu8FUT vV0iphek 9Vvye5xS q7rJJyxa yHiIEMHA Ce8KLI1B FdbpdvWY qLk23poI aRoZ5LTu fWNL8rcj RpZyI052 HTxj28Q0 GiOjJ1UN iW7zrxBD QPpkiBVE nvOAkh7p c2prdKB8 9DAYvYo5 BPSN8wmO Q2oNZouQ zfEjm5aC lLMDotic hi585ip4 c7LYN3LZ xGmpN32s lcF83ipK 0IwvvEe1 tQxKHCCa u51OKNIE kdEsXUHG tTpUtwbG T6E4hMYv nVpbxCPH 0aACMPtu Oq945xMi wlPQHJ1e bROJU0e7 wdBjAYPt gjIaTuLu bicVsgYN L3a5NLwf 30zu9OHL qtDs1PJM OmTsSOZc v4eM7s8f MQlppFcY 6HTWrZPZ Raj94J30 kcSQPdTQ zsOhnhCQ sQDQkA3a uBP00Du8 qoq7syqj urFj9bqQ TV1EDcpC 4jKGRY27 vb3KgZQy EJillDeB UN4YYoLI hWgf1kqn o1B5s6Wm 98fQL4W0 PXaQeRc2 E45QBYtr od4CfqUo YsPizANv WFJj0nhM h7maM5WQ HuDYldsX qy1NLYCZ ZkvkuCxI hcD6Hyod sDiFWy4n tElzo9YK NNdt31gx NaeEtqmR MGwCCYWu y80zQlGX OAYoTGVY wYs20iOY j4eZDalG HDcd6eWZ Wvxqh0RI jykQ3bNt qRjxSxt6 4HjBIMK1 AIX5UEPr 1HQKp2ZH Fie3kxjb tzwmAigF QntpzTJO 9jQiDIDE LD0OlrSk 8PfSKmt4 MQBr2cK0 FLUQLq2h JfmjaCYv DqkdKyr8 ZtGnI5rj iqhACPMu UsY6ZIpT NjjgMBPV RW4YRcnZ Gyr9nest 9tIXI0km plugRQRv AlFpi0PJ DLcM8Zoq Auk5RBWs tMpfMMlU p6jGYq3Z rTIBTHVM zGFwFwQi j4O1AY21 BJnaiScY`
	// match at the very beginning: the first 100 characters or less
	assert.Equal(t,
		"<b>VWwXetig</b> mty8fORN UNia4NFm SQsfyFHk BLDdgVnc AcvKP2fs q8KxPH1A IaCzFj96 J0S2fqca jp3ElV9f ULIZ1aMX …",
		snippets("VWwXetig", s))
	// the first 100 … the match, at most 50 (50 from the start of the match)
	assert.Equal(t,
		"VWwXetig mty8fORN UNia4NFm SQsfyFHk BLDdgVnc AcvKP2fs q8KxPH1A IaCzFj96 J0S2fqca jp3ElV9f ULIZ1aMX … <b>GSUO1pLI</b> p7vuJie8 kPfc0ONq EthfIUjm u74guCZ8 IiJYxlR6 …",
		snippets("GSUO1pLI", s))
	// the first 100 … less than 50, the match, at most 50
	assert.Equal(t,
		"VWwXetig mty8fORN UNia4NFm SQsfyFHk BLDdgVnc AcvKP2fs q8KxPH1A IaCzFj96 J0S2fqca jp3ElV9f ULIZ1aMX … GSUO1pLI p7vuJie8 <b>kPfc0ONq</b> EthfIUjm u74guCZ8 IiJYxlR6 5j5LlapY TGO98fOQ …",
		snippets("kPfc0ONq", s))
	// the first 100 … 50, the match, at most 50
	assert.Equal(t,
		"VWwXetig mty8fORN UNia4NFm SQsfyFHk BLDdgVnc AcvKP2fs q8KxPH1A IaCzFj96 J0S2fqca jp3ElV9f ULIZ1aMX … u74guCZ8 IiJYxlR6 5j5LlapY TGO98fOQ fO2RUb1g <b>W8zaPa0v</b> ps0haNzW OOeFwf1h 1N3td7zk 0OoMX8Ek aTd3Ciea …",
		snippets("W8zaPa0v", s))
	// match at the very end
	assert.Equal(t,
		"VWwXetig mty8fORN UNia4NFm SQsfyFHk BLDdgVnc AcvKP2fs q8KxPH1A IaCzFj96 J0S2fqca jp3ElV9f ULIZ1aMX … tMpfMMlU p6jGYq3Z rTIBTHVM zGFwFwQi j4O1AY21 <b>BJnaiScY</b>",
		snippets("BJnaiScY", s))
	// match near the end
	assert.Equal(t,
		"VWwXetig mty8fORN UNia4NFm SQsfyFHk BLDdgVnc AcvKP2fs q8KxPH1A IaCzFj96 J0S2fqca jp3ElV9f ULIZ1aMX … Auk5RBWs tMpfMMlU p6jGYq3Z rTIBTHVM zGFwFwQi <b>j4O1AY21</b> BJnaiScY",
		snippets("j4O1AY21", s))

}
