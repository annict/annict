# == Schema Information
#
# Table name: twitter_bots
#
#  id         :integer          not null, primary key
#  name       :string           not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  index_twitter_bots_on_name  (name) UNIQUE
#

class TwitterBot < ActiveRecord::Base
end
