# https://gist.github.com/kinopyo/1338738

module OmniauthMock
  rspec

  def mock_auth_hash(uid = '12345')
    hash = {
      provider: 'twitter',
      uid:      uid,
      info: {
        nickname:  'mockuser',
        image: 'https://pbs.twimg.com/profile_images/451369673401987072/d8LOz14r.png'
      },
      credentials: {
        token:  'mock_token',
        secret: 'mock_secret'
      }
    }

    # The mock_auth configuration allows you to set per-provider (or default)
    # authentication hashes to return during integration testing.
    OmniAuth.config.mock_auth[:twitter] = hash

    hash
  end

  OmniAuth.config.test_mode = true
end
