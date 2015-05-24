class EpisodeDecorator < Draper::Decorator
  include Diffable

  delegate_all

  def title_with_number
    if object.number.present?
      if object.title.present?
        "#{object.number}「#{object.title}」"
      else
        object.number
      end
    else
      object.title
    end
  end

  def to_diffable_resource
    hash = {}

    white_list = %w(number sort_number title next_episode_id)

    white_list.each do |column_name|
      hash[column_name] = get_diffable_episode(column_name, object.send(column_name))
    end

    hash
  end
end
