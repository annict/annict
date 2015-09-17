class UserSocialFriendsQuery
  def initialize(user)
    @user = user
  end

  def all
    if @user.authorized_to?(:twitter) && @user.authorized_to?(:facebook)
      twitter_and_facebook_users
    else
      twitter_or_facebook_users
    end
  end

  private

  def twitter_and_facebook_users
    twitter_uids =  TwitterService.new(@user).uids
    facebook_uids = FacebookService.new(@user).uids

    t = Provider.arel_table
    twitter_conds = t[:name].eq("twitter").and(t[:uid].in(twitter_uids))
    facebook_conds = t[:name].eq("facebook").and(t[:uid].in(facebook_uids))
    User.joins(:providers).where(twitter_conds.or(facebook_conds))
  end

  def twitter_or_facebook_users
    provider = @user.providers.first
    uids = case provider.name
      when "twitter" then TwitterService.new(@user).uids
      when "facebook" then FacebookService.new(@user).uids
    end

    User.joins(:providers).where(providers: { name: provider.name, uid: uids })
  end
end
