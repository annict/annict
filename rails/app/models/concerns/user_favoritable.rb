# typed: false
# frozen_string_literal: true

module UserFavoritable
  extend ActiveSupport::Concern

  included do
    def favorite?(resource)
      resource.users.exists?(id)
    end

    def favorite(resource)
      return if favorite?(resource)

      favorite_resource = resource.favorites.create(user: self)

      return if favorite_resource.instance_of?(CharacterFavorite)

      FavoritableWatchedWorksCountJob.perform_later(favorite_resource, self)
    end

    def unfavorite(resource)
      favorite = resource.favorites.find_by(user: self)
      favorite.destroy if favorite.present?
    end

    def update_watched_works_count
      person_favorites.each do |person_favorite|
        person_favorite.update_watched_works_count(self)
      end

      organization_favorites.each do |organization_favorite|
        organization_favorite.update_watched_works_count(self)
      end
    end
  end
end
