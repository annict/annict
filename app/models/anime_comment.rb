# frozen_string_literal: true
# == Schema Information
#
# Table name: anime_comments
#
#  id         :bigint           not null, primary key
#  body       :string           not null
#  locale     :string           default("other"), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#  anime_id   :bigint           not null
#  user_id    :bigint           not null
#
# Indexes
#
#  index_anime_comments_on_anime_id              (anime_id)
#  index_anime_comments_on_locale                (locale)
#  index_anime_comments_on_user_id               (user_id)
#  index_anime_comments_on_user_id_and_anime_id  (user_id,anime_id) UNIQUE
#
# Foreign Keys
#
#  fk_rails_...  (anime_id => animes.id)
#  fk_rails_...  (user_id => users.id)
#

class AnimeComment < ApplicationRecord
  belongs_to :user
  belongs_to :work

  validates :body, length: { maximum: 150 }, allow_blank: true
end
