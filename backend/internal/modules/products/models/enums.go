package models

type Origin string

const (
	OriginNacionalExceto3_4_5_8          Origin = "0 - Nacional, exceto as indicadas nos códigos 3, 4, 5 e 8"
	OriginEstrangeiraImportacaoDireta    Origin = "1 - Estrangeira, importação direta, exceto a indicada no código 6"
	OriginEstrangeiraMercadoInterno      Origin = "2 - Estrangeira, adquirida no mercado interno, exceto a indicada no código 7"
	OriginNacionalConteudoImport40_70    Origin = "3 - Nacional: mercadoria ou bem com conteúdo de importação superior a 40% e inferior ou igual a 70%"
	OriginNacionalProcessosProdutivos    Origin = "4 - Nacional, cuja produção tenha sido feita em conformidade com os processos produtivos básicos de que tratam o Decreto-Lei nº 288/1967, e as Leis nº 8.248/1991, 8.387/1991, 10.176/2001 e 11.484/2007"
	OriginNacionalConteudoImport40       Origin = "5 - Nacional: mercadoria ou bem com Conteúdo de Importação inferior ou igual a 40%"
	OriginEstrangeiraImportacaoDiretaSem Origin = "6 - Estrangeira: importação direta, sem similar nacional, constante em lista de Resolução Camex e gás natural"
	OriginEstrangeiraMercadoInternoSem   Origin = "7 - Estrangeira: adquirida no mercado interno, sem similar nacional, constante em lista de Resolução Camex e gás natural"
	OriginNacionalConteudoImport70       Origin = "8 - Nacional: mercadoria ou bem com Conteúdo de Importação superior a 70%"
)
