# == Schema Information
#
# Table name: programs
#
#  id             :integer          not null, primary key
#  channel_id     :integer          not null
#  episode_id     :integer          not null
#  work_id        :integer          not null
#  started_at     :datetime         not null
#  created_at     :datetime
#  updated_at     :datetime
#  sc_last_update :datetime
#
# Indexes
#
#  index_programs_on_channel_id_and_episode_id  (channel_id,episode_id) UNIQUE
#

class Program < ActiveRecord::Base
  belongs_to :channel
  belongs_to :episode
  belongs_to :work
end
