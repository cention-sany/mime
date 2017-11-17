package quotedprintable

import (
	"io/ioutil"
	"strings"
	"testing"
)

var tstData1 = []struct {
	in, out string
}{
	{ //1
		in:  "=C3=83=C2=B6rfr=C3=83=C2=A5",
		out: "=C3=83=C2=B6rfr=C3=83=C2=A5"},
	{ //2
		in:  "class=3D\"\">Ringv\xc3\r\n\xa4gen 14, ",
		out: "class=3D\"\">Ringv\xc3\xa4gen 14, "},
	{ //3
		in:  "\xC3\x85\xC3\x84\xC3\x96. \xC3\xA5\xC3\xA4\r\n\xC3\xB6.",
		out: "\xC3\x85\xC3\x84\xC3\x96. \xC3\xA5\xC3\xA4\r\n\xC3\xB6."},
	{ //4
		in:  "ara kostnad f\xC3\n\x83\xC2\xB6r order 768298",
		out: "ara kostnad f\xC3\x83\xC2\xB6r order 768298"},
	{ //5
		in:  "fru Susanne och jag \xC3\r\n\xB6nskar ",
		out: "fru Susanne och jag \xC3\xB6nskar "},
	{ //6
		in:  "language:EN-US\">Bra fr\xC3\n\xA5ga som jag",
		out: "language:EN-US\">Bra fr\xC3\xA5ga som jag"},
	{ //7 - only proper end of line pattern (\n or \r\n) will be filtered
		in:  "language:EN-US\">Bra fr\xC3\n\n\xA5ga som jag",
		out: "language:EN-US\">Bra fr\xC3\n\n\xA5ga som jag"},
	{ //8
		in:  "lang你\r\n好uage:EN-US\">test 刘\r健\n胜123",
		out: "lang你\r\n好uage:EN-US\">test 刘\r健\n胜123"},
	{ //9
		in:  "lang \r\rå\n\n, ä, ö\r\n好uage:EN-US\">test 刘\r å, ä\r, \nö\n胜123",
		out: "lang \r\rå\n\n, ä, ö\r\n好uage:EN-US\">test 刘\r å, ä\r, \nö\n胜123"},
	{ //10
		in:  "lang你\r\n好uage:EN-US\">test \xF0\x9F\r\xA4\x90 (f09fa490)123",
		out: "lang你\r\n好uage:EN-US\">test \xF0\x9F\r\xA4\x90 (f09fa490)123"},
	{ //11
		in:  "lang你\r\n好uage:EN-US\">test \xF0\n\x9F\xA4\x90 (f09fa490)123",
		out: "lang你\r\n好uage:EN-US\">test 🤐 (f09fa490)123"},
	{ //12
		in: `	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	A
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	B
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	C
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	E
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789🤐`,
		out: `	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	A
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	B
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	C
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	E
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789🤐`},
	{ //13
		in:  "leveransbekr\xC3\xA4ftelse\xE2\x80\n\x9D eller vad menas? <br>",
		out: "leveransbekr\xC3\xA4ftelse\xE2\x80\x9D eller vad menas? <br>"},
	{ //14
		in:  "leveransbekr\xC3\xA4ftelse\xE2\x80\r\n\x9D eller vad menas? <br>",
		out: "leveransbekr\xC3\xA4ftelse\xE2\x80\x9D eller vad menas? <br>"},
	{ //15 - only proper \n or \r\n will be filtered
		in:  "leveransbekr\xC3\xA4ftelse\xE2\x80\n\n\x9D eller vad menas? <br>",
		out: "leveransbekr\xC3\xA4ftelse\xE2\x80\n\n\x9D eller vad menas? <br>"},
}

func Test_qpUTF8(t *testing.T) {
	for i, d := range tstData1 {
		r := newQPUTF8(strings.NewReader(d.in))
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("[%d] FAILED expected nil error but got %v", i+1, err)
		}
		out := string(b)
		if out != d.out {
			t.Errorf("[%d] FAILED expected: %s but got %s", i+1, d.out, out)
		}
	}
}

var tstData2 = []struct {
	in, out string
}{
	{ //1
		in:  "<div class=3D\"\"><span style=3D\"FONT-SIZE: 10pt;mso-fareast-font-family: 'Ti=\r\nmes New Roman';\" class=3D\"\"></span><strong class=3D\"\"><span style=3D\"FONT-F=\r\nAMILY",
		out: "<div class=\"\"><span style=\"FONT-SIZE: 10pt;mso-fareast-font-family: 'Times New Roman';\" class=\"\"></span><strong class=\"\"><span style=\"FONT-FAMILY"},
	{ //2
		in:  "this messag=\r\ne",
		out: "this message"},
	{ //3
		in:  "\xC3\x85\xC3\x84\xC3\x96. \xC3\xA5\xC3\xA4\r\n\xC3\xB6.",
		out: "\xC3\x85\xC3\x84\xC3\x96. \xC3\xA5\xC3\xA4\r\n\xC3\xB6."},
	{ //4
		in:  "<p>Are you sure that these are the final prices that we receive?<o:p></o:p>=\r\n</p>",
		out: "<p>Are you sure that these are the final prices that we receive?<o:p></o:p></p>"},
	{ //5
		in:  "Denna f=C3=83=C2=B6rfr=C3=83=C2=A5gan g=C3=83=C2=A4ller bara kostnad f=C3\r\n=83=C2=B6r order 768298",
		out: "Denna f\xC3\x83\xC2\xB6rfr\xC3\x83\xC2\xA5gan g\xC3\x83\xC2\xA4ller bara kostnad f\xC3\x83\xC2\xB6r order 768298"},
	{ //6
		in:  "class=3D\"\">Ringv=C3\n=A4gen 14, SE-341",
		out: "class=\"\">Ringv\xC3\xA4gen 14, SE-341"},
	{ //7
		in:  "fru Susanne och jag =C3\r\n=B6nskar er en vacker dag",
		out: "fru Susanne och jag \xC3\xB6nskar er en vacker dag"},
	{ //8
		in:  "Bra fr=C3\r\n=A5ga som jag inte",
		out: "Bra fr\xC3\xA5ga som jag inte"},
	{ //9
		in:  "",
		out: ""},
	{ //10
		in:  "p=C3=83=C2=A5 Laggon =C3=83 =C2\r\n=A4r fel, jag kollade",
		out: "p\xC3\x83\xC2\xA5 Laggon \xC3\x83 \xC2\xA4r fel, jag kollade"},
	{ //11
		in:  "lang你\r\n好uage:EN-US\">test \xF0\n\x9F\xA4\x90 (f09fa490)123",
		out: "lang你\r\n好uage:EN-US\">test 🤐 (f09fa490)123"},
	{ //12
		in: `	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	A
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	B
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	C
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	E
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789🤐`,
		out: `	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	A
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	B
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	C
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	E
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
	23456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789🤐`},
	{ //13
		in:  "leveransbekr=C3=A4ftelse=E2=80\n=9D eller vad menas? <br>",
		out: "leveransbekr\xC3\xA4ftelse\xE2\x80\x9D eller vad menas? <br>"},
	{ //14
		in:  "leveransbekr=C3=A4ftelse=E2=80\r\n=9D eller vad menas? <br>",
		out: "leveransbekr\xC3\xA4ftelse\xE2\x80\x9D eller vad menas? <br>"},
	{ //15 - only proper \n or \r\n will be filtered
		in:  "leveransbekr=C3=A4ftelse=E2=80\n\r=9D eller vad menas? <br>",
		out: "leveransbekr\xC3\xA4ftelse\xE2\x80\n\r\x9D eller vad menas? <br>"},
}

func Test_NewUTF8Reader(t *testing.T) {
	for i, d := range tstData2 {
		r := NewUTF8Reader(strings.NewReader(d.in))
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("[%d] FAILED expected nil error but got %v", i+1, err)
		}
		out := string(b)
		if out != d.out {
			t.Errorf("[%d] FAILED expected: %s but got %s", i+1, d.out, out)
		}
	}
}