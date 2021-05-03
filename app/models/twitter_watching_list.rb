# frozen_string_literal: true

# == Schema Information
#
# Table name: twitter_watching_lists
#
#  id                  :bigint           not null, primary key
#  discord_webhook_url :string           not null
#  name                :string           not null
#  username            :string           not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#  since_id            :string
#

class TwitterWatchingList < ApplicationRecord
end
