* テーブル構成
  * userテーブル(Kratosが管理)
    * User ID
      * int
    * User Name
      * str
    * E-mail
      * str
    * password
      * str

  * VMテーブル（Redis）
    * User ID
      * int
      * userテーブル User IDの外部キー
    * VMs
      * int[]

