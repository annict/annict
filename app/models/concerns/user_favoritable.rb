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
      favorite_people.each do |favorite_person|
        favorite_person.update_watched_works_count(self)
      end

      favorite_organizations.each do |favorite_org|
        favorite_org.update_watched_works_count(self)
      end
    end
  end
end
