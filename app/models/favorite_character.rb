# frozen_string_literal: true
# == Schema Information
#
# Table name: favorite_characters
#
#  id           :integer          not null, primary key
#  user_id      :integer          not null
#  character_id :integer          not null
#  created_at   :datetime         not null
#  updated_at   :datetime         not null
#
# Indexes
#
#  index_favorite_characters_on_character_id              (character_id)
#  index_favorite_characters_on_user_id                   (user_id)
#  index_favorite_characters_on_user_id_and_character_id  (user_id,character_id) UNIQUE
#

class FavoriteCharacter < ApplicationRecord
  belongs_to :character
  belongs_to :user, counter_cache: true
end
