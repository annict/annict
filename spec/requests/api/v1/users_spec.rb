# frozen_string_literal: true

describe "Api::V1::Users" do
  describe "GET /v1/users" do
    let(:user) { create(:user, :with_profile) }
    let(:application) { create(:oauth_application, owner: user) }
    let(:access_token) { create(:oauth_access_token, owner: user, application: application) }

    before do
      Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
      get api("/v1/users", access_token: access_token.token)
    end

    after do
      Timecop.return
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets user info" do
        expected_hash = {
          "id" => user.id,
          "username" => user.username,
          "name" => user.profile.name,
          "description" => "悟空を倒すために生まれました。よろしくお願いします。",
          "url" => "http://example.com",
          "records_count" => 0,
          "created_at" => "2017-01-28T23:39:04.000Z"
        }
        expect(json["users"][0].stringify_keys).to include(expected_hash)
        expect(expected_hash).to include(json["users"][0].stringify_keys)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
