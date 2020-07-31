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
#  user_id    :bigint           not null
#  work_id    :bigint           not null
#
# Indexes
#
#  index_anime_comments_on_locale               (locale)
#  index_anime_comments_on_user_id              (user_id)
#  index_anime_comments_on_user_id_and_work_id  (user_id,work_id) UNIQUE
#  index_anime_comments_on_work_id              (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => animes.id)
#

class AnimeComment < ApplicationRecord
  belongs_to :user
  belongs_to :work

  validates :body, length: { maximum: 150 }, allow_blank: true
end
