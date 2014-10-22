class Program < ActiveRecord::Base
  belongs_to :channel
  belongs_to :episode
  belongs_to :work
end
