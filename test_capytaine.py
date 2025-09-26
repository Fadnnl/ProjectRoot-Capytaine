import numpy as np
from capytaine import BEMSolver, FloatingBody, RadiationProblem
from capytaine.meshes.meshes import Mesh

def test_simple_cube():
    # Buat mesh cube sederhana (8 titik, 6 face)
    vertices = [
        (-1, -1, -1),
        (-1, -1,  1),
        (-1,  1, -1),
        (-1,  1,  1),
        ( 1, -1, -1),
        ( 1, -1,  1),
        ( 1,  1, -1),
        ( 1,  1,  1),
    ]
    faces = [
        (0, 1, 3, 2),  # kiri
        (4, 5, 7, 6),  # kanan
        (0, 1, 5, 4),  # bawah
        (2, 3, 7, 6),  # atas
        (0, 2, 6, 4),  # depan
        (1, 3, 7, 5),  # belakang
    ]
    mesh = Mesh(vertices=vertices, faces=faces, name="cube")

    # Bungkus ke FloatingBody
    body = FloatingBody(mesh=mesh, name="cube")
    body.add_translation_dof(name="Heave", direction=(0, 0, 1))

    # Solver
    solver = BEMSolver()
    omega = 2 * np.pi * 1.0  # 1 Hz

    # Buat RadiationProblem
    problem = RadiationProblem(
        body=body,
        radiating_dof="Heave",
        omega=omega,
        water_depth=np.inf,
    )

    # Solve pakai problem object
    result = solver.solve(problem)

    # ✅ Assert disesuaikan
    assert "Heave" in result.added_masses
    print("=== Capytaine Cube Test OK ===")
    print("Added mass (Heave, 1Hz):", result.added_masses["Heave"])
