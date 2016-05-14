# == Schema Information
#
# Table name: providers
#
#  id               :integer          not null, primary key
#  user_id          :integer          not null
#  name             :string           not null
#  uid              :string           not null
#  token            :string           not null
#  token_expires_at :integer
#  token_secret     :string
#  created_at       :datetime
#  updated_at       :datetime
#
# Indexes
#
#  index_providers_on_name_and_uid  (name,uid) UNIQUE
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
