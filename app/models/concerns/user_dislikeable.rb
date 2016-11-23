# frozen_string_literal: true

module UserDislikeable
  extend ActiveSupport::Concern

  included do
    def dislike?(recipient)
      dislikes.where(recipient: recipient).present?
    end

    def dislike(recipient)
      dislikes.create(recipient: recipient) unless dislike?(recipient)
    end

    def undislike(recipient)
      dislike = dislikes.where(recipient: recipient).first
      dislike.destroy if dislike.present?
    end
  end
end
