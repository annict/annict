# frozen_string_literal: true

class FavoriteCharactersController < ApplicationV6Controller
  def index
    set_page_category PageCategory::FAVORITE_CHARACTER_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @favorite_characters = @user.favorite_characters.order(id: :desc)
  end
end
