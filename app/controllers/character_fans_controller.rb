# typed: false
# frozen_string_literal: true

class CharacterFansController < ApplicationV6Controller
  def index
    set_page_category PageCategory::CHARACTER_FAN_LIST

    @character = Character.only_kept.find(params[:character_id])
    @character_favorites = @character
      .character_favorites
      .joins(:user)
      .merge(User.only_kept)
      .order(id: :desc)
  end
end
