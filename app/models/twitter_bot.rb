# == Schema Information
#
# Table name: twitter_bots
#
#  id         :bigint           not null, primary key
#  name       :string(510)      not null
#  created_at :timestamptz
#  updated_at :timestamptz
#
# Indexes
#
#  twitter_bots_name_key  (name) UNIQUE
#

class TwitterBot < ApplicationRecord
end
