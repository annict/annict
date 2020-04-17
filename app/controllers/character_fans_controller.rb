# frozen_string_literal: true

class CharacterFansController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index
    @character = Character.only_kept.find(params[:character_id])
    @favorite_characters = @character.
      favorite_characters.
      joins(:user).
      merge(User.only_kept).
      order(id: :desc)
  end

  private

  def load_i18n
    keys = {
      "verb.follow": nil,
      "noun.following": nil,
      "messages._components.favorite_button.add_to_favorites": nil,
      "messages._components.favorite_button.added_to_favorites": nil
    }

    load_i18n_into_gon keys
  end
end
