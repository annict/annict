# typed: false
# frozen_string_literal: true

class CharacterFavorite < ApplicationRecord
  counter_culture :character, column_name: :favorite_users_count
  counter_culture :user

  belongs_to :character
  belongs_to :user
end
