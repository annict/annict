# frozen_string_literal: true

class EmailNotificationService
  def self.send_on_follow(user, following_user)
    new(user).send_following(following_user)
  end

  def initialize(user)
    @user = user
  end

  def send_on_follow(following_user)
    if @user.email_receivable?(:follow)
    end
  end
end
