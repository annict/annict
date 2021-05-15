# frozen_string_literal: true

module V3
  class FavoriteCharactersController < V3::ApplicationController
    before_action :load_i18n, only: %i[index]

    def index
      set_page_category PageCategory::FAVORITE_CHARACTER_LIST

      @user = User.only_kept.find_by!(username: params[:username])
      @user_entity = V4::UserEntity.from_model(@user)
      @favorite_characters = @user.favorite_characters.order(id: :desc)
    end

    private

    def load_i18n
      keys = {
        "messages._components.favorite_button.add_to_favorites": nil,
        "messages._components.favorite_button.added_to_favorites": nil
      }

      load_i18n_into_gon keys
    end
  end
end
