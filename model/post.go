package model

func likeFactory() {

}

/*LikeMess - сообщение для загрузки Like в память*/
type LikeMess struct {
	likes  []Like
	ltemps []LTemp
}

/*Temp - промежуточная структура*/
type LTemp struct {
	Liker int64 `json:"liker"`
	Likee int64 `json:"likee"`
	Ts    int64 `json:"ts"`
}
type LikesT struct {
	Likes []LTemp `json:"likes"`
}
