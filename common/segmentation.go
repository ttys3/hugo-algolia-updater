package common

import (
	"log"
	"strings"
	"sync/atomic"

	"github.com/deckarep/golang-set"
	"github.com/go-ego/gse"
	"github.com/yanyiwu/gojieba"
)

var (
	seg   gse.Segmenter
	jieba *gojieba.Jieba
)

func InitJieba() {
	dictPath := GetCurrentPath() + "/data/dict.txt"
	seg.LoadDict(dictPath)

	jiebaPathArray := strings.Split(dictPath, ",")
	jieba = gojieba.NewJieba(jiebaPathArray...)

	stopPath := GetCurrentPath() + "/data/stop.txt"

	stopStr := ReadFileString(stopPath)
	if stopStr != "" {
		StopArray = strings.Split(stopStr, "\n")
	} else {
		stop_str := "一,、,。,七,☆,〈,∈,〉,三,昉,《,》,「,」,『,』,‐,【,】,В,—,〔,―,∕,〕,‖,〖,〗,‘,’,“,”,〝,〞,!,\",•,#,$,%,&,…,',㈧,∧,(,),*,∪,+,,,-,.,/,︰,′,︳,″,︴,︵,︶,︷,︸,‹,︹,:,›,︺,;,︻,<,︼,=,︽,>,︾,?,︿,@,﹀,﹁,﹂,﹃,﹄,≈,义,﹉,﹊,﹋,﹌,﹍,﹎,﹏,﹐,﹑,﹔,﹕,﹖,[,\\,],九,﹝,^,﹞,_,﹟,也,`,﹠,①,﹡,②,﹢,③,④,﹤,⑤,⑥,﹦,⑦,⑧,﹨,⑨,﹩,⑩,﹪,﹫,|,白,~,二,五,¦,«,¯,±,´,·,¸,»,¿,ˇ,ˉ,ˊ,ˋ,×,四,˜,零,÷,─,！,＂,＃,℃,＄,％,＆,＇,（,）,＊,＋,，,－,．,／,0,１,２,３,４,５,６,７,８,９,：,会,；,＜,＝,＞,？,＠,Ａ,Ｂ,Ｃ,Ｄ,Ｅ,Ｆ,Ｇ,Ｉ,Ｌ,Ｒ,Ｔ,Ｘ,Ｚ,［,］,＿,ａ,ｂ,ｃ,ｄ,ｅ,ｆ,ｇ,ｈ,ｉ,ｊ,ｎ,ｏ,｛,｜,｝,～,Ⅲ,↑,→,Δ,■,Ψ,▲,β,γ,λ,μ,ξ,φ,￣,￥,\\n,},{,0,1,2,3,4,5,6,7,8,9,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z,\n,\t,\r, ,.."
		StopArray = strings.Split(stop_str, ",")
	}
}

func Participles(title string, content string) []string {
	jiebaArray := jieBaParticiples(content)
	segoArray := segoParticiples(content)
	jiebaSet := array2set(jiebaArray)
	segoSet := array2set(segoArray)
	set := segoSet.Union(jiebaSet)
	// set:=jiebaSet

	set = removeWord(set)
	slice := set.ToSlice()
	array := InterfaceArray2StringArray(slice)
	log.Printf("Participles ---------> title=%s array=%s", title, array)
	atomic.AddInt32(&Num, int32(len(array)))
	return array
}

func jieBaParticiples(context string) []string {
	// defer jieba.Free()
	jiebaParticiplesArray := jieba.CutForSearch(context, true)
	return jiebaParticiplesArray
}

func segoParticiples(context string) []string {
	return seg.CutAll(context)
}

// 接口数组转字符串数组
func InterfaceArray2StringArray(interfaceArray []interface{}) []string {
	var stringArray []string
	for _, str := range interfaceArray {
		if maybeStr, ok := str.(string); !ok {
			continue
		} else {
			// skip single word like 呢 / 吧 / 做
			if len([]rune(maybeStr)) < 2 {
				continue
			}
			stringArray = append(stringArray, maybeStr)
		}
	}
	return stringArray
}

// 接口数组转字符串数组
func array2set(aArray []string) mapset.Set {
	set := mapset.NewSet()
	for _, obj := range aArray {
		set.Add(obj)
	}

	return set
}

// 取出停顿词
func removeWord(wordSet mapset.Set) mapset.Set {
	for index := range StopArray {
		wordSet.Remove(StopArray[index])
	}
	return wordSet
}
