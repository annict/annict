class UserSocialFriendsQuery
  def initialize(user)
    @user = user
  end

  def all
    twitter_uids = TwitterService.new(@user).uids
    facebook_uids = FacebookService.new(@user).uids

    t = Provider.arel_table
    twitter_conds = t[:name].eq("twitter").and(t[:uid].in(twitter_uids))
    facebook_conds = t[:name].eq("facebook").and(t[:uid].in(facebook_uids))
    User.joins(:providers).where(twitter_conds.or(facebook_conds))
  end
end
