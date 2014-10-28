# == Schema Information
#
# Table name: providers
#
#  id               :integer          not null, primary key
#  user_id          :integer          not null
#  name             :string(510)      not null
#  uid              :string(510)      not null
#  token            :string(510)      not null
#  token_expires_at :integer
#  token_secret     :string(510)
#  created_at       :datetime
#  updated_at       :datetime
#
# Indexes
#
#  providers_name_uid_key  (name,uid) UNIQUE
#  providers_user_id_idx   (user_id)
#

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
