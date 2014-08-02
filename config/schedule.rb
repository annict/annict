every 1.day, at: '5:00 am' do
  rake 'syobocal:save'
end

every 1.day, at: '5:30 am' do
  rake 'channel_work:update'
end