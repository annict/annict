# frozen_string_literal: true

module UserLikeable
  extend ActiveSupport::Concern

  included do
    def like?(likeable)
      likes.where(likeable: likeable).present?
    end

    def like!(resource)
      unless resource.likeable?
        raise Annict::Errors::NotLikeableError
      end

      likes.create!(likeable: resource)
    end

    def unlike(likeable)
      like = likes.where(likeable: likeable).first
      like.destroy if like.present?
      like
    end
  end
end
