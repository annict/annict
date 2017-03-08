# frozen_string_literal: true

module UserFavoritable
  extend ActiveSupport::Concern

  included do
    def favorite?(resource)
      resource.users.exists?(self)
    end

    def favorite(resource)
      resource.favorites.create(user: self) unless favorite?(resource)
    end

    def unfavorite(resource)
      favorite = resource.favorites.find_by(user: self)
      favorite.destroy if favorite.present?
    end
  end
end
