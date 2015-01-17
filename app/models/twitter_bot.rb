# == Schema Information
#
# Table name: twitter_bots
#
#  id         :integer          not null, primary key
#  name       :string(510)      not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  twitter_bots_name_key  (name) UNIQUE
#

class TwitterBot < ActiveRecord::Base
end
