# frozen_string_literal: true

class Deprecated::UserSocialFriendsQuery
  def initialize(user)
    @user = user
  end

  def all
    if @user.authorized_to?(:twitter, shareable: true) && @user.authorized_to?(:facebook, shareable: true)
      twitter_and_facebook_users
    elsif @user.authorized_to?(:twitter, shareable: true) || @user.authorized_to?(:facebook, shareable: true)
      users_via(@user.providers.first.name)
    else
      User.none
    end
  end

  def users_via(provider_name)
    uids = case provider_name.to_s
    when "twitter" then Deprecated::TwitterService.new(@user).uids
    when "facebook" then Deprecated::FacebookService.new(@user).uids
    end

    User.only_kept.joins(:providers).where(providers: {name: provider_name, uid: uids})
  end

  private

  def twitter_and_facebook_users
    twitter_uids = Deprecated::TwitterService.new(@user).uids
    facebook_uids = Deprecated::FacebookService.new(@user).uids

    t = Provider.arel_table
    twitter_conds = t[:name].eq("twitter").and(t[:uid].in(twitter_uids))
    facebook_conds = t[:name].eq("facebook").and(t[:uid].in(facebook_uids))
    User.only_kept.joins(:providers).where(twitter_conds.or(facebook_conds))
  end
end
