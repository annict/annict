# frozen_string_literal: true

describe "API::Internal::Favorites" do
  let!(:user) { create(:registered_user) }

  describe "POST /api/internal/favorites/unfavorite" do
    before do
      login_as(user, scope: :user)
    end

    context "when user unfavorites character" do
      let!(:favorite_character) { create(:favorite_character, user: user) }
      let!(:data) { { resource_type: "Character", resource_id: favorite_character.character_id } }

      it "responses 200" do
        post api("/api/internal/favorites/unfavorite", data)
        expect(response.status).to eq(200)
      end

      it "removes a record" do
        expect { post api("/api/internal/favorites/unfavorite", data) }.
          to change { user.favorite_characters.count }.from(1).to(0)
      end
    end

    context "when user unfavorites organization" do
      let!(:favorite_organization) { create(:favorite_organization, user: user) }
      let!(:data) { { resource_type: "Organization", resource_id: favorite_organization.organization_id } }

      it "responses 200" do
        post api("/api/internal/favorites/unfavorite", data)
        expect(response.status).to eq(200)
      end

      it "removes a record" do
        expect { post api("/api/internal/favorites/unfavorite", data) }.
          to change { user.favorite_organizations.count }.from(1).to(0)
      end
    end

    context "when user unfavorites person" do
      let!(:favorite_person) { create(:favorite_person, user: user) }
      let!(:data) { { resource_type: "Person", resource_id: favorite_person.person_id } }

      it "responses 200" do
        post api("/api/internal/favorites/unfavorite", data)
        expect(response.status).to eq(200)
      end

      it "removes a record" do
        expect { post api("/api/internal/favorites/unfavorite", data) }.
          to change { user.favorite_people.count }.from(1).to(0)
      end
    end
  end
end
