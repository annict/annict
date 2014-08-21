class Provider < ActiveRecord::Base
  belongs_to :user


  def token_expires_at=(expires_at)
    expires_at if 'facebook' == name
  end
  
  def token_secret=(secret)
    secret if 'twitter' == name
  end
end
