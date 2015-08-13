class MultipleEpisodeValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, _value)
    record.to_episode_hash
  rescue
    record.errors[attribute] << (options[:message] || "が不正な値です。")
  end
end
