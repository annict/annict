# frozen_string_literal: true

describe "Api::Internal::Favorites" do
  let!(:user) { create(:registered_user) }

  describe "POST /api/internal/favorites/unfavorite" do
    before do
      login_as(user, scope: :user)
    end

    context "when user unfavorites character" do
      let!(:character_favorite) { create(:character_favorite, user: user) }
      let!(:data) { {resource_type: "Character", resource_id: character_favorite.character_id} }

      it "responses 200" do
        post api("/api/internal/favorites/unfavorite", data)
        expect(response.status).to eq(200)
      end

      it "removes a record" do
        expect { post api("/api/internal/favorites/unfavorite", data) }
          .to change { user.favorite_characters.count }.from(1).to(0)
      end
    end

    context "when user unfavorites organization" do
      let!(:organization_favorite) { create(:organization_favorite, user: user) }
      let!(:data) { {resource_type: "Organization", resource_id: organization_favorite.organization_id} }

      it "responses 200" do
        post api("/api/internal/favorites/unfavorite", data)
        expect(response.status).to eq(200)
      end

      it "removes a record" do
        expect { post api("/api/internal/favorites/unfavorite", data) }
          .to change { user.favorite_organizations.count }.from(1).to(0)
      end
    end

    context "when user unfavorites person" do
      let!(:person_favorite) { create(:person_favorite, user: user) }
      let!(:data) { {resource_type: "Person", resource_id: person_favorite.person_id} }

      it "responses 200" do
        post api("/api/internal/favorites/unfavorite", data)
        expect(response.status).to eq(200)
      end

      it "removes a record" do
        expect { post api("/api/internal/favorites/unfavorite", data) }
          .to change { user.favorite_people.count }.from(1).to(0)
      end
    end
  end
end
