class Provider < ActiveRecord::Base
  belongs_to :user


  def token_expires_at=(expires_at)
    value = expires_at if name == 'facebook'
    write_attribute(:token_expires_at, value)
  end

  def token_secret=(secret)
    value = secret if 'twitter' == name
    write_attribute(:token_secret, value)
  end
end
