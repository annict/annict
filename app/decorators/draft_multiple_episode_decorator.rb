class DraftMultipleEpisodeDecorator < ApplicationDecorator
  def to_values
    body = to_episode_hash.map do |episode|
      str = ""
      str += episode[:number].to_s
      str += "「#{episode[:title]}」" if episode[:title].present?
      str
    end.join("\n")

    { body: h.simple_format(body, {}, wrapper_tag: :div) }
  end
end
