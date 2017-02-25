# frozen_string_literal: true

describe "Api::V1::Me::Index" do
  describe "GET /v1/me" do
    let(:user) { create(:user, :with_profile) }
    let(:access_token) { create(:oauth_access_token, owner: user) }

    before do
      Timecop.freeze(Time.parse("2017-01-29 08:39:04"))

      data = {
        access_token: access_token.token
      }
      get api("/v1/me", data)
    end

    after do
      Timecop.return
    end

    it "responses 200" do
      expect(response.status).to eq(200)
    end

    it "gets user" do
      expected_hash = {
        "id" => user.id,
        "username" => user.username,
        "name" => user.profile.name,
        "description" => "悟空を倒すために生まれました。よろしくお願いします。",
        "url" => "http://example.com",
        "records_count" => 0,
        "created_at" => "2017-01-28T23:39:04.000Z",
        "email" => "#{user.username}@example.com",
        "notifications_count" => 0
      }
      expect(json.stringify_keys).to include(expected_hash)
      expect(expected_hash).to include(json.stringify_keys)
    end
  end
end
