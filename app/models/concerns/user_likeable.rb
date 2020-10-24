# frozen_string_literal: true

module UserLikeable
  extend ActiveSupport::Concern

  included do
    def like?(recipient)
      likes.where(recipient: recipient).present?
    end

    def like(recipient)
      likes.create(recipient: recipient) unless like?(recipient)
    end

    def unlike(recipient)
      like = likes.where(recipient: recipient).first
      like.destroy if like.present?
      like
    end
  end
end
