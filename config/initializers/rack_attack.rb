# frozen_string_literal: true

def remote_ip(req)
  @remote_ip ||= (req.env["action_dispatch.remote_ip"] || req.ip).to_s
end

Rack::Attack.throttle("requests by ip", limit: 5, period: 2.seconds) do |req|
  remote_ip(req)
end

ActiveSupport::Notifications.subscribe(/rack_attack/) do |_name, _start, _finish, _request_id, payload|
  req = payload[:request]

  next unless %i(throttle blacklist).include? req.env["rack.attack.match_type"]

  Rails.logger.info("Rate limit hit (#{req.env['rack.attack.match_type']}): #{remote_ip(req)} #{req.request_method} #{req.fullpath}")
end
