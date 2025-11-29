# typed: false

module UserFollowable
  extend ActiveSupport::Concern

  included do
    def following?(user)
      followings.where(id: user.id).present?
    end

    def followers
      User.includes(:profile).joins(:follows).where("follows.following_id": id)
    end

    def follow(user)
      follows.create(following: user) unless following?(user)
    end

    def unfollow(user)
      following = follows.where(following_id: user.id).first
      following.destroy if following.present?
    end
  end
end
