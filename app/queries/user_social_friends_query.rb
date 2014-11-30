class UserSocialFriendsQuery
  def initialize(user)
    @user = user
  end

  def all
    provider = @user.providers.first
    uids = case provider.name
      when 'twitter'  then TwitterService.new(@user).uids
      when 'facebook' then FacebookService.new(@user).uids
      end

    User.joins(:providers).where(providers: { name: provider.name, uid: uids })
  end
end
