# frozen_string_literal: true

module UserLikeable
  extend ActiveSupport::Concern

  included do
    def like?(recipient)
      likes.where(recipient: recipient).present?
    end

    def like(recipient)
      like = likes.find_by(recipient: recipient)

      if like
        return like
      end

      like = likes.create(recipient: recipient)
      like.send_notification_to(self)

      like
    end

    def unlike(recipient)
      like = likes.where(recipient: recipient).first
      like.destroy if like.present?
      like
    end
  end
end
