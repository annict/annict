class UserSocialFriendsQuery
  def initialize(user)
    @user = user
  end

  def all
    provider = @user.providers.first
    uids = case provider.name
      when 'twitter'  then @user.twitter_uids
      when 'facebook' then @user.facebook_uids
      end

    User.joins(:providers).where(providers: { name: provider.name, uid: uids })
  end
end
