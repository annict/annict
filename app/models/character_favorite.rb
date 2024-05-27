# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: character_favorites
#
#  id           :bigint           not null, primary key
#  created_at   :datetime         not null
#  updated_at   :datetime         not null
#  character_id :bigint           not null
#  user_id      :bigint           not null
#
# Indexes
#
#  index_character_favorites_on_character_id              (character_id)
#  index_character_favorites_on_user_id                   (user_id)
#  index_character_favorites_on_user_id_and_character_id  (user_id,character_id) UNIQUE
#
# Foreign Keys
#
#  fk_rails_...  (character_id => characters.id)
#  fk_rails_...  (user_id => users.id)
#

class CharacterFavorite < ApplicationRecord
  counter_culture :character, column_name: :favorite_users_count
  counter_culture :user

  belongs_to :character
  belongs_to :user
end
