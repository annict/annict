# frozen_string_literal: true
# == Schema Information
#
# Table name: twitter_watching_lists
#
#  id                  :bigint(8)        not null, primary key
#  username            :string           not null
#  name                :string           not null
#  since_id            :string
#  discord_webhook_url :string           not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#

class TwitterWatchingList < ApplicationRecord
end
