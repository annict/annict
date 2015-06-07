class MultipleEpisodeValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    episodes = record.to_episode_hash
  rescue
    record.errors[attribute] << (options[:message] || "が不正な値です。")
  end
end
