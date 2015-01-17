# == Schema Information
#
# Table name: items
#
#  id         :integer          not null, primary key
#  work_id    :integer
#  name       :string           not null
#  url        :string           not null
#  image_uid  :string           not null
#  main       :boolean          default("false"), not null
#  created_at :datetime
#  updated_at :datetime
#

class Item < ActiveRecord::Base
  dragonfly_accessor :image

  belongs_to :work
end
