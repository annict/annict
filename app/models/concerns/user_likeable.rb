# frozen_string_literal: true

module UserLikeable
  extend ActiveSupport::Concern

  included do
    def like?(recipient)
      likes.where(recipient: recipient).present?
    end

    def like!(resource)
      unless resource.likeable?
        raise Annict::Errors::NotLikeableError
      end

      recipient = case resource
      when Record
        resource.episode_record? ? resource.episode_record : resource.anime_record
      else
        resource
      end

      likes.create!(recipient: recipient)
    end

    def unlike(resource)
      recipient = case resource
      when Record
        resource.episode_record? ? resource.episode_record : resource.anime_record
      else
        resource
      end

      like = likes.where(recipient: recipient).first
      like.destroy if like.present?
      like
    end
  end
end
