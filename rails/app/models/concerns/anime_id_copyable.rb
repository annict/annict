# typed: false
# frozen_string_literal: true

# 作品・エピソードを指す行に、指す作品・エピソードのanime_idを保存する前に写す。
# 関連テーブルの行をanimeから直接引けるようにするための列で、値はwork_id / episode_idから導く。
# 写し損ねた行や既存の行は、Go版のanimes同期バッチが照合して直す。
module AnimeIdCopyable
  extend ActiveSupport::Concern

  class_methods do
    # 写し元の関連 (belongs_to) と写す先の列を宣言する。
    #
    # Annict DBの保存は検証を飛ばして保存する (save!(validate: false)) ため、検証のコールバックではなく
    # before_saveで写す。写し元の外部キーが変わらない保存では写さず、変更の無い保存で活動履歴に
    # 空の差分を作らないようにする。
    def copy_anime_id_from(association, to: :anime_id)
      foreign_key = reflect_on_association(association).foreign_key

      before_save do
        if will_save_change_to_attribute?(foreign_key)
          self[to] = public_send(association)&.anime_id
        end
      end
    end
  end
end
